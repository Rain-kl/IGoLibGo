// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import {
  AlertTriangle,
  Info,
  RefreshCw,
  Terminal,
  XCircle,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Spinner } from '@/components/ui/spinner';
import type { ActivityLogEntry } from '@/lib/services/igo/types';

interface ActivityLogStreamProps {
  logs: ActivityLogEntry[];
  loading: boolean;
  onRefresh: () => void;
  maxHeight?: string;
}

export function ActivityLogStream({
  logs,
  loading,
  onRefresh,
  maxHeight = 'max-h-80',
}: ActivityLogStreamProps) {
  const t = useTranslations('igo.dashboard');
  const tCommon = useTranslations('common');

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <Terminal className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('logsTitle')}
            </CardTitle>
            <Badge variant='outline' className='text-[10px] font-mono'>
              {logs.length} 条记录
            </Badge>
          </div>
          <Button
            variant='ghost'
            size='icon'
            aria-label='刷新日志'
            onClick={onRefresh}
            disabled={loading}
            className='size-7'
          >
            <RefreshCw
              className={`size-3.5 ${loading ? 'animate-spin' : ''}`}
            />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {loading && logs.length === 0 ? (
          <div className='py-12 flex flex-col items-center justify-center gap-2 text-xs text-muted-foreground'>
            <Spinner className='size-5' />
            <span>{tCommon('loading')}</span>
          </div>
        ) : logs.length === 0 ? (
          <div className='py-8 text-center text-xs text-muted-foreground border-dashed border rounded-lg'>
            {t('emptyLogs')}
          </div>
        ) : (
          <div
            className={`border rounded-lg bg-muted/20 font-mono text-xs overflow-y-auto ${maxHeight}`}
          >
            <div className='divide-y divide-border/40'>
              {logs.map((log) => {
                const isError = log.level === 'error';
                const isWarn = log.level === 'warn';

                return (
                  <div
                    key={log.id}
                    className='p-2 px-3 flex items-start gap-2 hover:bg-muted/40 transition-colors'
                  >
                    <div className='mt-0.5 shrink-0'>
                      {isError ? (
                        <XCircle className='size-3 text-destructive' />
                      ) : isWarn ? (
                        <AlertTriangle className='size-3 text-amber-500' />
                      ) : (
                        <Info className='size-3 text-primary/70' />
                      )}
                    </div>
                    <div className='shrink-0 text-[10px] text-muted-foreground/80 font-mono pt-0.5'>
                      {log.created_at
                        ? log.created_at.slice(11, 19)
                        : '--:--:--'}
                    </div>
                    {log.kind && (
                      <Badge
                        variant='outline'
                        className='text-[9px] py-0 px-1 h-4 shrink-0'
                      >
                        {log.kind}
                      </Badge>
                    )}
                    <div className='flex-1 break-all text-[11px] text-foreground leading-relaxed'>
                      {log.message}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
