// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { History } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import type { TaskLaunchRecord } from '@/lib/services/igo/types';

interface TomorrowHistoryTableProps {
  records: TaskLaunchRecord[];
}

export function TomorrowHistoryTable({ records }: TomorrowHistoryTableProps) {
  const t = useTranslations('igo.tomorrow');

  return (
    <Card className='border-dashed shadow-none'>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-2'>
          <History className='size-4 text-primary' />
          <CardTitle className='text-base font-semibold'>
            {t('historyTitle')}
          </CardTitle>
          <Badge variant='outline' className='text-[10px] font-mono'>
            {records.length} 条记录
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        {records.length === 0 ? (
          <div className='py-8 text-center text-xs text-muted-foreground border-dashed border rounded-lg'>
            {t('historyEmpty')}
          </div>
        ) : (
          <div className='border border-dashed shadow-none rounded-lg overflow-hidden bg-background'>
            <Table className='text-xs'>
              <TableHeader className='bg-muted/40'>
                <TableRow className='border-dashed hover:bg-transparent'>
                  <TableHead className='w-40'>记录时间</TableHead>
                  <TableHead>场馆</TableHead>
                  <TableHead>目标座位</TableHead>
                  <TableHead className='w-24 text-right'>状态</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {records.map((rec) => (
                  <TableRow
                    key={rec.record_id}
                    className='border-dashed hover:bg-muted/10 transition-colors'
                  >
                    <TableCell className='font-mono text-muted-foreground text-[11px]'>
                      {rec.recorded_at}
                    </TableCell>
                    <TableCell className='font-medium'>
                      {rec.library_name || `场馆 #${rec.library_id}`}
                    </TableCell>
                    <TableCell className='font-mono'>
                      {rec.seats
                        ?.map((s) => s.seat_name || s.seat_key)
                        .join(', ') || '--'}
                    </TableCell>
                    <TableCell className='text-right'>
                      <Badge
                        variant='outline'
                        className='text-[10px] font-medium rounded-full py-0 px-2'
                      >
                        已归档
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
