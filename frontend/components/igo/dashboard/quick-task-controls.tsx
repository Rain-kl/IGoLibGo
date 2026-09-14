// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import {
  ArrowUpRight,
  Calendar,
  Clock,
  Compass,
  Radio,
  Rocket,
  ShieldCheck,
  Zap,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { TaskStatusBadge } from '@/components/igo/common/task-status-badge';
import type { CoordinatorStatus } from '@/lib/services/igo/types';

interface QuickTaskControlsProps {
  tasks: CoordinatorStatus[];
  loading: boolean;
}

interface EngineDefinition {
  kind: string;
  name: string;
  url: string;
  icon: React.ComponentType<{ className?: string }>;
  description: string;
}

const ENGINES: EngineDefinition[] = [
  {
    kind: 'grab',
    name: '抢座引擎',
    url: '/grab',
    icon: Rocket,
    description: '毫秒级高频冲刺抢座与多候选轮询',
  },
  {
    kind: 'leak',
    name: '全域捡漏',
    url: '/leak',
    icon: Radio,
    description: '全馆退座/违规释放实时监听捡漏',
  },
  {
    kind: 'tomorrow',
    name: '明日预约',
    url: '/tomorrow',
    icon: Calendar,
    description: '准点定时次日预约与提前预热',
  },
  {
    kind: 'occupy',
    name: '占座防暂离',
    url: '/occupy',
    icon: ShieldCheck,
    description: '自动续期守护防暂离超时违规释放',
  },
];

export function QuickTaskControls({ tasks, loading }: QuickTaskControlsProps) {
  const t = useTranslations('igo.dashboard');

  const taskMap = React.useMemo(() => {
    const map = new Map<string, CoordinatorStatus>();
    for (const task of tasks) {
      map.set(task.kind, task);
    }
    return map;
  }, [tasks]);

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-2'>
          <Zap className='size-4 text-primary' />
          <CardTitle className='text-base font-semibold'>
            {t('quickControls')}
          </CardTitle>
        </div>
      </CardHeader>
      <CardContent>
        <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3'>
          {ENGINES.map((engine) => {
            const task = taskMap.get(engine.kind);
            const state = task?.state || 'idle';
            const Icon = engine.icon;

            return (
              <div
                key={engine.kind}
                className='p-3.5 rounded-lg border bg-card hover:border-primary/40 hover:bg-muted/30 transition-all flex flex-col justify-between space-y-3'
              >
                <div className='space-y-2'>
                  <div className='flex items-center justify-between'>
                    <div className='flex items-center gap-2 font-semibold text-sm text-foreground'>
                      <Icon className='size-4 text-primary' />
                      <span>{engine.name}</span>
                    </div>
                    <TaskStatusBadge state={state} />
                  </div>
                  <p className='text-[11px] text-muted-foreground line-clamp-2'>
                    {task?.message || engine.description}
                  </p>
                </div>

                <div className='pt-2 border-t border-border/40 flex items-center justify-between text-[11px] text-muted-foreground'>
                  <span className='font-mono text-[10px]'>
                    {task?.last_request_at
                      ? `请求于 ${task.last_request_at.slice(11, 19)}`
                      : '待触发'}
                  </span>
                  <Button
                    variant='ghost'
                    size='sm'
                    asChild
                    className='h-6 px-2 text-xs gap-0.5 text-primary'
                  >
                    <Link href={engine.url}>
                      <span>进入</span>
                      <ArrowUpRight className='size-3' />
                    </Link>
                  </Button>
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
