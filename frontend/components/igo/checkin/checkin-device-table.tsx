// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { GraduationCap, Radio, Smartphone, User } from 'lucide-react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { CheckInDeviceResponse } from '@/lib/services/igo/types';

interface CheckInDeviceTableProps {
  device: CheckInDeviceResponse | null;
  loading: boolean;
}

export function CheckInDeviceTable({
  device,
  loading,
}: CheckInDeviceTableProps) {
  const t = useTranslations('igo.checkin');

  if (!device && !loading) {
    return (
      <Card className='border-dashed shadow-none'>
        <CardContent className='py-8 text-center text-xs text-muted-foreground'>
          {t('noDevices')}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-2'>
          <Smartphone className='size-4 text-primary' />
          <CardTitle className='text-base font-semibold'>
            {t('deviceListTitle')}
          </CardTitle>
        </div>
        <CardDescription>
          从微信端获取的绑定学籍与当前场馆允许的蓝牙 Beacon 广播参数
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        {device ? (
          <>
            <div className='grid grid-cols-1 sm:grid-cols-3 gap-3 p-3 rounded-lg bg-muted/40 text-xs'>
              <div className='flex items-center gap-2'>
                <User className='size-3.5 text-muted-foreground' />
                <div>
                  <div className='text-muted-foreground'>姓名与学号</div>
                  <div className='font-semibold text-foreground'>
                    {device.student_name || device.nickname} (
                    {device.student_number || '--'})
                  </div>
                </div>
              </div>
              <div className='flex items-center gap-2'>
                <GraduationCap className='size-3.5 text-muted-foreground' />
                <div>
                  <div className='text-muted-foreground'>所属学校</div>
                  <div className='font-medium text-foreground truncate'>
                    {device.school || '--'}
                  </div>
                </div>
              </div>
              <div className='flex items-center gap-2'>
                <Radio className='size-3.5 text-muted-foreground' />
                <div>
                  <div className='text-muted-foreground'>
                    可用 Beacon UUID 数量
                  </div>
                  <div className='font-mono font-semibold text-primary'>
                    {device.beacon_uuids?.length ?? 0} 个
                  </div>
                </div>
              </div>
            </div>

            <div className='space-y-1.5'>
              <div className='text-xs font-medium text-foreground flex items-center gap-1.5'>
                <Radio className='size-3.5 text-primary' />
                <span>可用 Beacon UUID 列表</span>
              </div>
              <div className='flex flex-wrap gap-1.5 max-h-24 overflow-y-auto p-2 border rounded-lg bg-muted/15 font-mono text-[11px]'>
                {device.beacon_uuids?.map((uuid) => (
                  <Badge
                    key={uuid}
                    variant='outline'
                    className='text-[10px] font-mono py-0.5'
                  >
                    {uuid}
                  </Badge>
                )) || (
                  <span className='text-xs text-muted-foreground'>
                    暂无可用 Beacon
                  </span>
                )}
              </div>
            </div>
          </>
        ) : (
          <div className='text-xs text-muted-foreground py-2'>
            正在加载授权设备信息...
          </div>
        )}
      </CardContent>
    </Card>
  );
}
