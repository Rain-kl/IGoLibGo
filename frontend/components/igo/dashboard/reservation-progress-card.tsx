// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import {
  AlertTriangle,
  Building2,
  CalendarCheck,
  Clock,
  ExternalLink,
  MapPin,
  RefreshCw,
  XCircle,
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
import { Progress } from '@/components/ui/progress';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { IGoService } from '@/lib/services/igo';
import type { ReservationResponse } from '@/lib/services/igo/types';

interface ReservationProgressCardProps {
  reservation: ReservationResponse | null;
  loading: boolean;
  onRefresh: () => void;
}

export function ReservationProgressCard({
  reservation,
  loading,
  onRefresh,
}: ReservationProgressCardProps) {
  const t = useTranslations('igo.dashboard');
  const tCommon = useTranslations('common');

  const [cancelOpen, setCancelOpen] = React.useState(false);
  const [cancelling, setCancelling] = React.useState(false);
  const [refreshing, setRefreshing] = React.useState(false);

  const hasReservation = reservation?.has_reservation ?? false;

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await IGoService.reservation.refreshReservation();
      toast.success(tCommon('save'));
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setRefreshing(false);
    }
  };

  const handleCancel = async () => {
    setCancelling(true);
    try {
      await IGoService.reservation.cancelReservation({
        stop_occupy_first: true,
      });
      toast.success(tCommon('save'));
      setCancelOpen(false);
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setCancelling(false);
    }
  };

  return (
    <>
      <Card className='border-border/60 shadow-sm'>
        <CardHeader className='pb-3'>
          <div className='flex items-center justify-between'>
            <div className='flex items-center gap-2'>
              <CalendarCheck className='size-4 text-primary' />
              <CardTitle className='text-base font-semibold'>
                {t('reservationTitle')}
              </CardTitle>
            </div>
            <div className='flex items-center gap-2'>
              {hasReservation ? (
                <Badge
                  variant='outline'
                  className='border-emerald-500/40 text-emerald-600 dark:text-emerald-400 text-xs'
                >
                  在座使用中
                </Badge>
              ) : (
                <Badge variant='secondary' className='text-xs'>
                  暂无在座
                </Badge>
              )}
              <Button
                variant='ghost'
                size='icon'
                aria-label='刷新预约'
                onClick={handleRefresh}
                disabled={refreshing || loading}
                className='size-7'
              >
                <RefreshCw
                  className={`size-3.5 ${refreshing ? 'animate-spin' : ''}`}
                />
              </Button>
            </div>
          </div>
          <CardDescription>
            {hasReservation
              ? `${reservation?.library_name || '场馆'} · ${reservation?.seat_name || '座位'}`
              : t('noReservation')}
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          {hasReservation && reservation ? (
            <div className='space-y-3'>
              <div className='grid grid-cols-1 sm:grid-cols-3 gap-3 p-3 rounded-lg bg-muted/40 text-xs'>
                <div className='flex items-center gap-2'>
                  <Building2 className='size-3.5 text-muted-foreground' />
                  <div>
                    <div className='text-muted-foreground'>
                      {t('reservationAt')}
                    </div>
                    <div className='font-semibold text-foreground truncate'>
                      {reservation.library_name || '--'}
                    </div>
                  </div>
                </div>
                <div className='flex items-center gap-2'>
                  <MapPin className='size-3.5 text-muted-foreground' />
                  <div>
                    <div className='text-muted-foreground'>{t('seatAt')}</div>
                    <div className='font-mono font-semibold text-foreground'>
                      {reservation.seat_name || reservation.seat_key || '--'}
                    </div>
                  </div>
                </div>
                <div className='flex items-center gap-2'>
                  <Clock className='size-3.5 text-muted-foreground' />
                  <div>
                    <div className='text-muted-foreground'>
                      {t('expiresAt')}
                    </div>
                    <div className='font-mono font-semibold text-primary'>
                      {reservation.expiration_time || '--'}
                    </div>
                  </div>
                </div>
              </div>

              <div className='flex items-center justify-between pt-1'>
                <Button
                  variant='outline'
                  size='sm'
                  asChild
                  className='h-8 text-xs gap-1'
                >
                  <Link href='/occupy'>
                    <span>前往续座守护</span>
                    <ExternalLink className='size-3' />
                  </Link>
                </Button>

                <Button
                  variant='destructive'
                  size='sm'
                  onClick={() => setCancelOpen(true)}
                  className='h-8 text-xs gap-1'
                >
                  <XCircle className='size-3.5' />
                  <span>{t('cancelReservation')}</span>
                </Button>
              </div>
            </div>
          ) : (
            <div className='p-6 text-center border-dashed border rounded-lg space-y-2'>
              <div className='text-xs text-muted-foreground'>
                {t('noReservation')}
              </div>
              <div className='flex justify-center gap-2 pt-1'>
                <Button
                  variant='outline'
                  size='sm'
                  asChild
                  className='h-7 text-xs'
                >
                  <Link href='/venue'>选座预约</Link>
                </Button>
                <Button
                  variant='default'
                  size='sm'
                  asChild
                  className='h-7 text-xs'
                >
                  <Link href='/grab'>启动抢座</Link>
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 取消确认弹窗 */}
      <AlertDialog open={cancelOpen} onOpenChange={setCancelOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className='flex items-center gap-2 text-destructive'>
              <AlertTriangle className='size-5' />
              {t('cancelConfirmTitle')}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t('cancelConfirmDesc')}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={cancelling}>
              {tCommon('cancel')}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancel}
              disabled={cancelling}
              className='bg-destructive hover:bg-destructive/90 text-destructive-foreground'
            >
              {cancelling ? '正在释放...' : tCommon('confirm')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
