// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import {
  CheckCircle2,
  ClipboardPaste,
  KeyRound,
  LogOut,
  QrCode,
  RefreshCw,
  ShieldCheck,
  UserCheck,
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
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Spinner } from '@/components/ui/spinner';
import { IGoService } from '@/lib/services/igo';
import type {
  CheckInSessionResponse,
  QRCodeResponse,
} from '@/lib/services/igo/types';

interface CheckInAuthCardProps {
  session: CheckInSessionResponse | null;
  loading: boolean;
  onRefresh: () => void;
}

export function CheckInAuthCard({
  session,
  loading,
  onRefresh,
}: CheckInAuthCardProps) {
  const t = useTranslations('igo.checkin');
  const tCommon = useTranslations('common');

  // 微信二维码扫码弹窗状态
  const [qrOpen, setQrOpen] = React.useState(false);
  const [qrData, setQrData] = React.useState<QRCodeResponse | null>(null);
  const [qrLoading, setQrLoading] = React.useState(false);
  const [qrLinkInput, setQrLinkInput] = React.useState('');
  const [rememberQr, setRememberQr] = React.useState(true);
  const [submittingQr, setSubmittingQr] = React.useState(false);

  // 手动录入弹窗状态
  const [manualOpen, setManualOpen] = React.useState(false);
  const [manualInput, setManualInput] = React.useState('');
  const [rememberManual, setRememberManual] = React.useState(true);
  const [submittingManual, setSubmittingManual] = React.useState(false);

  const handleOpenQr = async () => {
    setQrOpen(true);
    setQrLoading(true);
    try {
      const res = await IGoService.checkin.getAuthQRCode();
      setQrData(res);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('loadFailed'));
    } finally {
      setQrLoading(false);
    }
  };

  const handlePasteTo = async (setter: (val: string) => void) => {
    try {
      if (typeof navigator !== 'undefined' && navigator.clipboard?.readText) {
        const text = await navigator.clipboard.readText();
        if (text) {
          setter(text.trim());
          toast.success(t('pasteSuccess'));
        }
      }
    } catch {
      // 剪贴板权限或环境受限时静默处理，用户可手动 Ctrl+V / Cmd+V
    }
  };

  const handleSubmitCode = async (
    rawInput: string,
    remember: boolean,
    isFromQr: boolean,
  ) => {
    const trimmed = rawInput.trim();
    if (!trimmed) return;
    if (isFromQr) {
      setSubmittingQr(true);
    } else {
      setSubmittingManual(true);
    }
    try {
      await IGoService.checkin.authorizeFromCode({
        code: trimmed,
        remember,
      });
      toast.success(t('authSuccess'));
      if (isFromQr) {
        setQrOpen(false);
        setQrLinkInput('');
      } else {
        setManualOpen(false);
        setManualInput('');
      }
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      if (isFromQr) {
        setSubmittingQr(false);
      } else {
        setSubmittingManual(false);
      }
    }
  };

  const handleClearSession = async () => {
    try {
      await IGoService.checkin.clearSession();
      toast.success(tCommon('save'));
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
              <UserCheck className='size-4 text-primary' />
              <CardTitle className='text-base font-semibold'>
                {t('authTitle')}
              </CardTitle>
            </div>
            {loading ? (
              <Spinner className='size-4' />
            ) : isAuthorized ? (
              <Badge
                variant='outline'
                className='gap-1 border-emerald-500/40 text-emerald-600 dark:text-emerald-400'
              >
                <ShieldCheck className='size-3' />
                <span>{t('authorized')}</span>
              </Badge>
            ) : (
              <Badge variant='secondary' className='text-muted-foreground'>
                {t('unauthorized')}
              </Badge>
            )}
          </div>
          <CardDescription>{t('description')}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          {isAuthorized && session ? (
            <div className='grid grid-cols-1 sm:grid-cols-2 gap-3 p-3 rounded-lg bg-muted/40 text-xs'>
              <div>
                <span className='text-muted-foreground'>授权保存时间：</span>
                <span className='font-mono font-medium text-foreground'>
                  {session.saved_at || '--'}
                </span>
              </div>
              <div>
                <span className='text-muted-foreground'>到期时间：</span>
                <span className='font-mono font-medium text-foreground'>
                  {session.expires_at || '--'}
                </span>
              </div>
            </div>
          ) : (
            <div className='p-3 rounded-lg bg-muted/40 text-xs text-muted-foreground'>
              {t('waitingAuth')}
            </div>
          )}

          <div className='flex items-center gap-2 pt-1'>
            <Button
              variant='outline'
              size='sm'
              onClick={handleOpenQr}
              className='gap-1.5 border-dashed shadow-none'
            >
              <QrCode className='size-3.5' />
              <span>{t('scanQrBtn')}</span>
            </Button>
            <Button
              variant='outline'
              size='sm'
              onClick={() => setManualOpen(true)}
              className='gap-1.5 border-dashed shadow-none'
            >
              <KeyRound className='size-3.5' />
              <span>{t('manualInputBtn')}</span>
            </Button>
            {isAuthorized && (
              <Button
                variant='ghost'
                size='sm'
                onClick={handleClearSession}
                className='gap-1 text-destructive hover:bg-destructive/10 ml-auto'
              >
                <LogOut className='size-3.5' />
                <span>{t('clearSession')}</span>
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 签到微信扫码弹窗 */}
      <Dialog open={qrOpen} onOpenChange={setQrOpen}>
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle className='flex items-center gap-2'>
              <QrCode className='size-5 text-primary' />
              <span>{t('scanQrBtn')}</span>
            </DialogTitle>
            <DialogDescription>{t('manualDialogDesc')}</DialogDescription>
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
                  alt='CheckIn QR'
                  className='size-48 object-contain'
                />
              </div>
            ) : (
              <div className='py-8 text-center text-xs text-muted-foreground'>
                {t('refreshQr')}
              </div>
            )}

            <div className='flex items-start gap-2 p-2.5 rounded-lg bg-primary/5 border border-dashed border-primary/20 text-xs text-muted-foreground w-full'>
              <CheckCircle2 className='size-3.5 text-primary shrink-0 mt-0.5' />
              <span>{t('qrLinkTip')}</span>
            </div>

            <div className='w-full space-y-2 pt-2 border-t border-dashed'>
              <div className='flex items-center justify-between'>
                <Label
                  htmlFor='checkin-qr-link'
                  className='text-xs font-medium text-foreground'
                >
                  {t('qrLinkInputLabel')}
                </Label>
                <Button
                  type='button'
                  variant='ghost'
                  size='sm'
                  onClick={() => handlePasteTo(setQrLinkInput)}
                  className='h-6 px-2 text-xs text-muted-foreground hover:text-foreground gap-1'
                >
                  <ClipboardPaste className='size-3' />
                  <span>{t('pasteClipboard')}</span>
                </Button>
              </div>
              <Textarea
                id='checkin-qr-link'
                value={qrLinkInput}
                onChange={(e) => setQrLinkInput(e.target.value)}
                placeholder={t('qrLinkPlaceholder')}
                rows={2}
                className='font-mono text-xs shadow-none border-dashed break-all'
              />
              <div className='flex items-center justify-between pt-1'>
                <div className='flex items-center space-x-2'>
                  <Checkbox
                    id='remember-checkin-qr'
                    checked={rememberQr}
                    onCheckedChange={(c) => setRememberQr(!!c)}
                  />
                  <Label
                    htmlFor='remember-checkin-qr'
                    className='text-xs font-normal cursor-pointer text-muted-foreground'
                  >
                    {t('rememberSession')}
                  </Label>
                </div>
                <Button
                  size='sm'
                  onClick={() =>
                    handleSubmitCode(qrLinkInput, rememberQr, true)
                  }
                  disabled={!qrLinkInput.trim() || submittingQr}
                  className='gap-1.5 shadow-none'
                >
                  {submittingQr ? (
                    <Spinner className='size-3.5' />
                  ) : (
                    <KeyRound className='size-3.5' />
                  )}
                  <span>{t('parseAndLoginBtn')}</span>
                </Button>
              </div>
            </div>
          </div>
          <DialogFooter className='sm:justify-between'>
            <Button
              variant='outline'
              size='sm'
              onClick={handleOpenQr}
              disabled={qrLoading}
              className='gap-1 border-dashed shadow-none'
            >
              <RefreshCw
                className={`size-3.5 ${qrLoading ? 'animate-spin' : ''}`}
              />
              <span>{t('refreshQr')}</span>
            </Button>
            <Button
              variant='secondary'
              size='sm'
              onClick={() => setQrOpen(false)}
              className='shadow-none'
            >
              {tCommon('cancel')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 手动录入链接 / 授权码弹窗 */}
      <Dialog open={manualOpen} onOpenChange={setManualOpen}>
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle className='flex items-center gap-2'>
              <KeyRound className='size-5 text-primary' />
              <span>{t('manualDialogTitle')}</span>
            </DialogTitle>
            <DialogDescription>{t('manualDialogDesc')}</DialogDescription>
          </DialogHeader>
          <div className='space-y-4 py-2'>
            <div className='space-y-2'>
              <div className='flex items-center justify-between'>
                <Label
                  htmlFor='checkin-manual-input'
                  className='text-xs font-medium text-foreground'
                >
                  {t('qrLinkInputLabel')}
                </Label>
                <Button
                  type='button'
                  variant='ghost'
                  size='sm'
                  onClick={() => handlePasteTo(setManualInput)}
                  className='h-6 px-2 text-xs text-muted-foreground hover:text-foreground gap-1'
                >
                  <ClipboardPaste className='size-3' />
                  <span>{t('pasteClipboard')}</span>
                </Button>
              </div>
              <Textarea
                id='checkin-manual-input'
                value={manualInput}
                onChange={(e) => setManualInput(e.target.value)}
                placeholder={t('qrLinkPlaceholder')}
                rows={3}
                className='font-mono text-xs shadow-none border-dashed break-all'
              />
            </div>
            <div className='flex items-center space-x-2'>
              <Checkbox
                id='remember-checkin-manual'
                checked={rememberManual}
                onCheckedChange={(c) => setRememberManual(!!c)}
              />
              <Label
                htmlFor='remember-checkin-manual'
                className='text-xs font-normal cursor-pointer text-muted-foreground'
              >
                {t('rememberSession')}
              </Label>
            </div>
          </div>
          <DialogFooter className='sm:justify-between'>
            <Button
              variant='outline'
              size='sm'
              onClick={() => setManualOpen(false)}
              className='shadow-none border-dashed'
            >
              {tCommon('cancel')}
            </Button>
            <Button
              size='sm'
              onClick={() =>
                handleSubmitCode(manualInput, rememberManual, false)
              }
              disabled={!manualInput.trim() || submittingManual}
              className='gap-1.5 shadow-none'
            >
              {submittingManual ? (
                <Spinner className='size-3.5' />
              ) : (
                <KeyRound className='size-3.5' />
              )}
              <span>{t('parseAndLoginBtn')}</span>
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
