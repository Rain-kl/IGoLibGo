// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Sliders } from 'lucide-react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface OccupyIntervalConfigProps {
  intervalMode: string;
  onIntervalModeChange: (val: string) => void;
  reReserveDelaySeconds: number;
  onReReserveDelaySecondsChange: (val: number) => void;
}

export function OccupyIntervalConfig({
  intervalMode,
  onIntervalModeChange,
  reReserveDelaySeconds,
  onReReserveDelaySecondsChange,
}: OccupyIntervalConfigProps) {
  const t = useTranslations('igo.occupy');

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-2'>
          <Sliders className='size-4 text-primary' />
          <CardTitle className='text-base font-semibold'>
            续座与心跳策略
          </CardTitle>
        </div>
        <CardDescription>
          配置续座检查频率模式与提前发起重预约的缓冲时间
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
          <div className='space-y-1.5'>
            <Label className='text-xs font-medium'>
              {t('intervalModeTitle')}
            </Label>
            <Select value={intervalMode} onValueChange={onIntervalModeChange}>
              <SelectTrigger className='h-9 text-xs'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='fixed_10s' className='text-xs'>
                  {t('intervalFixed')}
                </SelectItem>
                <SelectItem value='random_10_20s' className='text-xs'>
                  {t('intervalRandom')}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='delay-seconds' className='text-xs font-medium'>
              {t('reReserveDelay')}
            </Label>
            <Input
              id='delay-seconds'
              type='number'
              value={reReserveDelaySeconds}
              onChange={(e) =>
                onReReserveDelaySecondsChange(
                  Math.max(10, Number(e.target.value)),
                )
              }
              min={10}
              max={600}
              className='h-9 text-xs font-mono'
            />
            <p className='text-[10px] text-muted-foreground'>
              {t('reReserveDelayHelp')}（推荐 60 秒）
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
