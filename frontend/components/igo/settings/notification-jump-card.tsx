// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { ArrowUpRight, Bell, MessagesSquare } from 'lucide-react';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';

export function NotificationJumpCard() {
  const t = useTranslations('igo.settings');

  return (
    <Card className='border-dashed shadow-none'>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-2'>
          <Bell className='size-4 text-primary' />
          <CardTitle className='text-base font-semibold'>
            {t('tabNotifications')}
          </CardTitle>
        </div>
        <CardDescription>Wavelet 统一通知中心与多通道消息网关</CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <p className='text-xs text-muted-foreground leading-relaxed'>
          {t('notificationsNotice')}
        </p>

        <div className='flex flex-wrap items-center gap-3 pt-1'>
          <Button
            variant='outline'
            size='sm'
            asChild
            className='gap-1 text-xs border-dashed shadow-none'
          >
            <Link href='/admin/push'>
              <Bell className='size-3.5 text-primary' />
              <span>{t('jumpToPushBtn')}</span>
              <ArrowUpRight className='size-3 text-muted-foreground' />
            </Link>
          </Button>

          <Button
            variant='outline'
            size='sm'
            asChild
            className='gap-1 text-xs border-dashed shadow-none'
          >
            <Link href='/admin/message-gateway'>
              <MessagesSquare className='size-3.5 text-primary' />
              <span>{t('jumpToGatewayBtn')}</span>
              <ArrowUpRight className='size-3 text-muted-foreground' />
            </Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
