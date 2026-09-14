// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import {
  CheckCircle2,
  KeyRound,
  LogOut,
  QrCode,
  RefreshCw,
  ShieldAlert,
  ShieldCheck,
} from 'lucide-react';
import { toast } from 'sonner';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';
import { IGoService } from '@/lib/services/igo';
import type { QRCodeResponse, SessionResponse } from '@/lib/services/igo/types';

interface SessionAuthCardProps {
  session: SessionResponse | null;
  loading: boolean;
  onRefresh: () => void;
}

export function SessionAuthCard({
  session,
  loading,
  onRefresh,
}: SessionAuthCardProps) {
  const t = useTranslations('igo.venue');
  const tCommon = useTranslations('common');

  // 二维码弹窗状态
  const [qrOpen, setQrOpen] = React.useState(false);
  const [qrData, setQrData] = React.useState<QRCodeResponse | null>(null);
  const [qrLoading, setQrLoading] = React.useState(false);
  const [qrLinkInput, setQrLinkInput] = React.useState('');
  const [rememberQrLink, setRememberQrLink] = React.useState(true);
  const [submittingQrLink, setSubmittingQrLink] = React.useState(false);

  // 手动 Cookie 录入弹窗状态
  const [cookieOpen, setCookieOpen] = React.useState(false);
  const [cookieInput, setCookieInput] = React.useState('');
  const [rememberCookie, setRememberCookie] = React.useState(true);
  const [submittingCookie, setSubmittingCookie] = React.useState(false);

  // 刷新会话中
  const [refreshingSession, setRefreshingSession] = React.useState(false);

  // 加载二维码
  const handleOpenQrDialog = async () => {
    setQrOpen(true);
    setQrLoading(true);
    try {
      const res = await IGoService.session.getAuthQRCode();
      setQrData(res);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('loadFailed'));
    } finally {
      setQrLoading(false);
    }
  };

  // 提交扫码后复制的链接
  const handleSubmitQrLink = async () => {
    if (!qrLinkInput.trim()) return;
    setSubmittingQrLink(true);
    try {
      await IGoService.session.authenticateFromCode({
        code: qrLinkInput.trim(),
        remember: rememberQrLink,
      });
      toast.success(tCommon('save'));
      setQrOpen(false);
      setQrLinkInput('');
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSubmittingQrLink(false);
    }
  };

  // 提交 Cookie 或授权链接
  const handleSubmitCookie = async () => {
    const trimmed = cookieInput.trim();
    if (!trimmed) return;
    setSubmittingCookie(true);
    try {
      const hasCode = /code=[A-Za-z0-9]{32}/i.test(trimmed);
      const isUrl =
        trimmed.startsWith('http://') || trimmed.startsWith('https://');
      const isRawCode = /^[A-Za-z0-9]{32}$/.test(trimmed);

      if (
        (hasCode || isUrl || isRawCode) &&
        !trimmed.toLowerCase().includes('authorization=')
      ) {
        await IGoService.session.authenticateFromCode({
          code: trimmed,
          remember: rememberCookie,
        });
      } else {
        await IGoService.session.authenticateFromCookie({
          cookie: trimmed,
          remember: rememberCookie,
        });
      }
      toast.success(tCommon('save'));
      setCookieOpen(false);
      setCookieInput('');
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSubmittingCookie(false);
    }
  };

  // 刷新 Cookie
  const handleRefreshCookie = async () => {
    setRefreshingSession(true);
    try {
      await IGoService.session.refreshCookie();
      toast.success(tCommon('save'));
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setRefreshingSession(false);
    }
  };

  // 登出断开
  const handleSignOut = async () => {
    try {
      await IGoService.session.signOut();
      toast.success(t('logout'));
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    }
  };

  const isAuthorized = session?.authorized ?? false;

  return (
    <>
      <Card className='border-dashed shadow-none'>
        <CardHeader className='pb-3'>
          <div className='flex items-center justify-between'>
            <div className='flex items-center gap-2'>
              <KeyRound className='size-4 text-primary' />
              <CardTitle className='text-base font-semibold'>
                {t('sessionCardTitle')}
              </CardTitle>
            </div>
            {loading ? (
              <Spinner className='size-4' />
            ) : isAuthorized ? (
              <Badge variant='default' className='gap-1'>
                <ShieldCheck className='size-3' />
                {t('authorized')}
              </Badge>
            ) : (
              <Badge
                variant='secondary'
                className='gap-1 text-muted-foreground'
              >
                <ShieldAlert className='size-3' />
                {t('unauthorized')}
              </Badge>
            )}
          </div>
          <CardDescription>{t('description')}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          {isAuthorized && session ? (
            <div className='grid grid-cols-1 sm:grid-cols-3 gap-3 p-3 rounded-lg bg-muted/40 text-xs'>
              <div>
                <span className='text-muted-foreground'>
                  {t('cookieSource')}:{' '}
                </span>
                <span className='font-medium text-foreground'>
                  {session.source || 'Session'}
                </span>
              </div>
              <div>
                <span className='text-muted-foreground'>
                  {t('expiresAt')}:{' '}
                </span>
                <span className='font-medium text-foreground'>
                  {session.expires_at || '--'}
                </span>
              </div>
              <div className='truncate'>
                <span className='text-muted-foreground'>
                  {t('cookieMasked')}:{' '}
                </span>
                <span className='font-mono font-medium text-foreground'>
                  {session.cookie_masked || '******'}
                </span>
              </div>
            </div>
          ) : (
            <div className='p-3 rounded-lg bg-muted/40 text-xs text-muted-foreground'>
              {t('waitingAuthDesc')}
            </div>
          )}

          <div className='flex flex-wrap items-center gap-2 pt-1'>
            <Button
              variant='outline'
              size='sm'
              onClick={handleOpenQrDialog}
              className='gap-1.5 border-dashed shadow-none'
            >
              <QrCode className='size-3.5' />
              {t('qrLoginBtn')}
            </Button>
            <Button
              variant='outline'
              size='sm'
              onClick={() => setCookieOpen(true)}
              className='gap-1.5 border-dashed shadow-none'
            >
              <KeyRound className='size-3.5' />
              {t('cookieLoginBtn')}
            </Button>
            {isAuthorized && (
              <>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={handleRefreshCookie}
                  disabled={refreshingSession}
                  className='gap-1.5 border-dashed shadow-none'
                >
                  <RefreshCw
                    className={`size-3.5 ${refreshingSession ? 'animate-spin' : ''}`}
                  />
                  {t('refreshCookie')}
                </Button>
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={handleSignOut}
                  className='gap-1.5 text-destructive hover:text-destructive hover:bg-destructive/10 ml-auto'
                >
                  <LogOut className='size-3.5' />
                  {t('logout')}
                </Button>
              </>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 微信二维码授权弹窗 */}
      <Dialog open={qrOpen} onOpenChange={setQrOpen}>
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle className='flex items-center gap-2'>
              <QrCode className='size-5 text-primary' />
              {t('qrDialogTitle')}
            </DialogTitle>
            <DialogDescription>{t('qrDialogDesc')}</DialogDescription>
          </DialogHeader>
          <div className='flex flex-col items-center justify-center p-4 space-y-4'>
            {qrLoading ? (
              <div className='py-12 flex flex-col items-center justify-center gap-2'>
                <Spinner className='size-8' />
                <p className='text-xs text-muted-foreground'>
                  {tCommon('loading')}
                </p>
              </div>
            ) : qrData?.image_data_url ? (
              <div className='p-2 bg-white rounded-lg border border-dashed shadow-none'>
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={qrData.image_data_url}
                  alt='TraceInt Auth QR'
                  className='size-48 object-contain'
                />
              </div>
            ) : (
              <div className='py-8 text-center text-sm text-muted-foreground'>
                {t('qrExpiring')}
              </div>
            )}
            <div className='flex items-center gap-2 text-xs text-muted-foreground'>
              <CheckCircle2 className='size-3.5 text-primary' />
              <span>
                微信扫码并授权后，可将跳转后的页面链接复制粘贴到下方快速绑定
              </span>
            </div>
            <div className='w-full space-y-2 pt-2 border-t border-dashed'>
              <Label className='text-xs font-medium text-foreground'>
                {t('qrLinkInputLabel')}
              </Label>
              <Textarea
                value={qrLinkInput}
                onChange={(e) => setQrLinkInput(e.target.value)}
                placeholder={t('qrLinkPlaceholder')}
                rows={2}
                className='font-mono text-xs shadow-none border-dashed break-all'
              />
              <div className='flex items-center justify-between pt-1'>
                <div className='flex items-center space-x-2'>
                  <Checkbox
                    id='remember-qr-session'
                    checked={rememberQrLink}
                    onCheckedChange={(c) => setRememberQrLink(!!c)}
                  />
                  <Label
                    htmlFor='remember-qr-session'
                    className='text-xs font-normal cursor-pointer text-muted-foreground'
                  >
                    {t('rememberSession')}
                  </Label>
                </div>
                <Button
                  size='sm'
                  onClick={handleSubmitQrLink}
                  disabled={!qrLinkInput.trim() || submittingQrLink}
                  className='gap-1.5 shadow-none'
                >
                  {submittingQrLink && <Spinner className='size-3.5' />}
                  {t('parseAndLoginBtn')}
                </Button>
              </div>
            </div>
          </div>
          <DialogFooter className='sm:justify-between'>
            <Button
              variant='outline'
              size='sm'
              onClick={handleOpenQrDialog}
              disabled={qrLoading}
              className='gap-1 border-dashed shadow-none'
            >
              <RefreshCw
                className={`size-3.5 ${qrLoading ? 'animate-spin' : ''}`}
              />
              {t('qrRefresh')}
            </Button>
            <Button
              variant='secondary'
              size='sm'
              onClick={() => {
                setQrOpen(false);
                onRefresh();
              }}
              className='shadow-none'
            >
              {tCommon('confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 手动 Cookie 录入弹窗 */}
      <Dialog open={cookieOpen} onOpenChange={setCookieOpen}>
        <DialogContent className='sm:max-w-lg'>
          <DialogHeader>
            <DialogTitle className='flex items-center gap-2'>
              <KeyRound className='size-5 text-primary' />
              {t('cookieDialogTitle')}
            </DialogTitle>
            <DialogDescription>{t('cookieDialogDesc')}</DialogDescription>
          </DialogHeader>
          <div className='space-y-4 py-2'>
            <Textarea
              value={cookieInput}
              onChange={(e) => setCookieInput(e.target.value)}
              placeholder={t('cookiePlaceholder')}
              rows={5}
              className='font-mono text-xs shadow-none border-dashed break-all whitespace-pre-wrap max-h-60 overflow-y-auto'
            />
            <div className='flex items-center space-x-2'>
              <Checkbox
                id='remember-session'
                checked={rememberCookie}
                onCheckedChange={(c) => setRememberCookie(!!c)}
              />
              <Label
                htmlFor='remember-session'
                className='text-xs font-normal cursor-pointer text-muted-foreground'
              >
                {t('rememberSession')}
              </Label>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant='outline'
              size='sm'
              onClick={() => setCookieOpen(false)}
              className='border-dashed shadow-none'
            >
              {tCommon('cancel')}
            </Button>
            <Button
              size='sm'
              onClick={handleSubmitCookie}
              disabled={!cookieInput.trim() || submittingCookie}
              className='gap-1.5 shadow-none'
            >
              {submittingCookie && <Spinner className='size-3.5' />}
              {tCommon('confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
