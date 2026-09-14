// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { AlertCircle, ArrowUpRight } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface GrabStrategyReminderProps {
  seatCount: number;
}

export function GrabStrategyReminder({ seatCount }: GrabStrategyReminderProps) {
  const t = useTranslations('igo.grab');

  if (seatCount >= 5) {
    return null;
  }

  return (
    <div className='flex items-start gap-2.5 p-3 rounded-lg border border-amber-500/30 bg-amber-500/10 text-xs text-foreground'>
      <AlertCircle className='size-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5' />
      <div className='flex-1 space-y-1'>
        <div className='font-medium text-amber-900 dark:text-amber-200'>
          最优抢座策略提示
        </div>
        <p className='text-muted-foreground leading-relaxed'>
          {t('strategyReminder', { count: seatCount })}
        </p>
      </div>
      <Button
        variant='ghost'
        size='sm'
        asChild
        className='h-7 text-xs text-primary shrink-0 gap-0.5'
      >
        <Link href='/venue'>
          <span>去选座</span>
          <ArrowUpRight className='size-3' />
        </Link>
      </Button>
    </div>
  );
}
