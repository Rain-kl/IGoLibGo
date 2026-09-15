// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import {
  AlertCircle,
  Building2,
  CheckCircle2,
  Clock,
  Radio,
  XCircle,
} from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import type { PipelineExecutionResult } from '@/lib/services/igo/types';
import { cn } from '@/lib/utils';

interface PipelineResultDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  result: PipelineExecutionResult | null;
}

export function PipelineResultDialog({
  open,
  onOpenChange,
  result,
}: PipelineResultDialogProps) {
  const t = useTranslations('igo.pipeline');

  if (!result) return null;

  const isSuccess = result.success;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-md p-0 gap-0 overflow-hidden'>
        <DialogHeader
          className={cn(
            'px-6 pt-6 pb-4 border-b border-border/40',
            isSuccess
              ? 'bg-emerald-500/10 dark:bg-emerald-950/20'
              : 'bg-rose-500/10 dark:bg-rose-950/20',
          )}
        >
          <div className='flex items-center gap-3'>
            <div
              className={cn(
                'flex size-10 shrink-0 items-center justify-center rounded-xl',
                isSuccess
                  ? 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-400'
                  : 'bg-rose-500/20 text-rose-600 dark:text-rose-400',
              )}
            >
              {isSuccess ? (
                <CheckCircle2 className='size-6' />
              ) : (
                <XCircle className='size-6' />
              )}
            </div>
            <div>
              <DialogTitle className='text-base font-semibold'>
                {isSuccess
                  ? t('resultDialog.successTitle')
                  : t('resultDialog.failureTitle')}
              </DialogTitle>
              <p className='text-xs text-muted-foreground mt-0.5'>
                配置:{' '}
                {result.name
                  ? `${result.name} (${result.config_id})`
                  : result.config_id}
              </p>
            </div>
          </div>
        </DialogHeader>

        <div className='px-6 py-5 space-y-3.5 text-sm'>
          {/* Detailed summary items */}
          <div className='rounded-lg bg-muted/30 border border-border/40 p-3.5 space-y-2.5'>
            {result.reservation_status && (
              <div className='flex items-center justify-between text-xs'>
                <span className='text-muted-foreground flex items-center gap-1.5'>
                  <Building2 className='size-3.5' />
                  {t('resultDialog.reserveResult')}
                </span>
                <span className='font-semibold'>
                  {result.reservation_status}
                </span>
              </div>
            )}

            {result.checkin_status && (
              <div className='flex items-center justify-between text-xs'>
                <span className='text-muted-foreground flex items-center gap-1.5'>
                  <Radio className='size-3.5' />
                  {t('resultDialog.checkinResult')}
                </span>
                <span className='font-semibold'>{result.checkin_status}</span>
              </div>
            )}

            <div className='flex items-center justify-between text-xs'>
              <span className='text-muted-foreground flex items-center gap-1.5'>
                <Clock className='size-3.5' />
                {t('resultDialog.executedAt')}
              </span>
              <span className='font-mono text-muted-foreground'>
                {result.executed_at}
              </span>
            </div>
          </div>

          {/* Feedback message */}
          {result.message && (
            <div
              className={cn(
                'rounded-lg p-3 text-xs leading-relaxed border',
                isSuccess
                  ? 'bg-emerald-500/5 text-emerald-800 dark:text-emerald-300 border-emerald-500/20'
                  : 'bg-rose-500/5 text-rose-800 dark:text-rose-300 border-rose-500/20',
              )}
            >
              <div className='flex items-start gap-2'>
                <AlertCircle className='size-4 shrink-0 mt-0.5' />
                <div className='flex-1 break-all'>{result.message}</div>
              </div>
            </div>
          )}
        </div>

        <DialogFooter className='px-6 py-3 border-t border-border/40 bg-muted/10 flex justify-end'>
          <Button
            size='sm'
            onClick={() => onOpenChange(false)}
            className='min-w-20'
          >
            {t('resultDialog.close')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
