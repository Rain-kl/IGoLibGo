// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import {
  AlertCircle,
  CheckCircle2,
  Loader2,
  PauseCircle,
  PlayCircle,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';

interface TaskStatusBadgeProps {
  state: string;
  className?: string;
}

export function TaskStatusBadge({ state, className }: TaskStatusBadgeProps) {
  const normalized = (state || 'idle').toLowerCase();

  switch (normalized) {
    case 'running':
    case 'starting':
      return (
        <Badge
          variant='default'
          className={`gap-1 bg-primary text-primary-foreground animate-pulse shadow-xs ${className || ''}`}
        >
          <Loader2 className='size-3 animate-spin' />
          <span>运行中</span>
        </Badge>
      );
    case 'stopping':
      return (
        <Badge
          variant='outline'
          className={`gap-1 border-amber-500/50 text-amber-600 dark:text-amber-400 ${className || ''}`}
        >
          <PauseCircle className='size-3' />
          <span>正在停止</span>
        </Badge>
      );
    case 'completed':
    case 'success':
      return (
        <Badge
          variant='outline'
          className={`gap-1 border-emerald-500/50 text-emerald-600 dark:text-emerald-400 ${className || ''}`}
        >
          <CheckCircle2 className='size-3' />
          <span>已完成</span>
        </Badge>
      );
    case 'failed':
    case 'error':
      return (
        <Badge variant='destructive' className={`gap-1 ${className || ''}`}>
          <AlertCircle className='size-3' />
          <span>异常/失败</span>
        </Badge>
      );
    default:
      return (
        <Badge
          variant='secondary'
          className={`gap-1 text-muted-foreground ${className || ''}`}
        >
          <PlayCircle className='size-3' />
          <span>待命中</span>
        </Badge>
      );
  }
}
