// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import {
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

  const [qrOpen, setQrOpen] = React.useState(false);
  const [qrData, setQrData] = React.useState<QRCodeResponse | null>(null);
  const [qrLoading, setQrLoading] = React.useState(false);

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

      {/* 签到扫码弹窗 */}
      <Dialog open={qrOpen} onOpenChange={setQrOpen}>
        <DialogContent className='sm:max-w-md'>
          <DialogHeader>
            <DialogTitle className='flex items-center gap-2'>
              <QrCode className='size-5 text-primary' />
              <span>微信扫码授权打卡</span>
            </DialogTitle>
            <DialogDescription>
              请使用微信扫描下方二维码，授权绑定用于远程打卡的微信身份：
            </DialogDescription>
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
                二维码失效，请点击刷新
              </div>
            )}
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
    </>
  );
}
