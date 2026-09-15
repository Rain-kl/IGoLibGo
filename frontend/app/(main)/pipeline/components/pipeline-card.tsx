// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import {
  Building2,
  CheckCircle2,
  Edit2,
  MapPin,
  MoreVertical,
  Play,
  Radio,
  ShieldAlert,
  ShieldCheck,
  Trash2,
  Workflow,
} from 'lucide-react';
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Spinner } from '@/components/ui/spinner';
import type { PipelineConfigDTO } from '@/lib/services/igo/types';
import { cn } from '@/lib/utils';

interface PipelineCardProps {
  config: PipelineConfigDTO;
  onRun: (config: PipelineConfigDTO) => void;
  onEdit: (config: PipelineConfigDTO) => void;
  onDelete: (config: PipelineConfigDTO) => void;
  isRunning?: boolean;
}

export function PipelineCard({
  config,
  onRun,
  onEdit,
  onDelete,
  isRunning,
}: PipelineCardProps) {
  const t = useTranslations('igo.pipeline');
  const tCommon = useTranslations('common');

  const isCookieValid = Boolean(
    config.has_cookie || config.cookie || config.cookie_masked,
  );
  const isCheckinValid =
    !config.auto_checkin ||
    Boolean(config.has_checkin_token || config.checkin_token);

  return (
    <Card className='relative flex flex-col justify-between overflow-hidden border border-border/60 bg-gradient-to-br from-card via-card/80 to-muted/20 transition-all duration-200 hover:shadow-md hover:border-primary/40'>
      <CardHeader className='pb-3 pt-5 px-5'>
        <div className='flex items-start justify-between gap-3'>
          <div className='flex items-center gap-2.5 min-w-0 flex-1'>
            <div className='flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary'>
              <Workflow className='size-5' />
            </div>
            <div className='min-w-0 flex-1'>
              <div className='flex items-center gap-2 flex-wrap'>
                <h2 className='text-base font-semibold tracking-tight truncate'>
                  {config.name}
                </h2>
                <Badge
                  variant='outline'
                  className='font-mono text-[11px] px-1.5 py-0 bg-muted/60'
                >
                  {config.id}
                </Badge>
              </div>
              <p className='text-xs text-muted-foreground mt-0.5'>
                {t('card.lastUpdated')}{' '}
                {new Date(config.updated_at).toLocaleDateString()}
              </p>
            </div>
          </div>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant='ghost'
                size='icon'
                aria-label={tCommon('actions')}
                className='size-8 text-muted-foreground hover:text-foreground'
              >
                <MoreVertical className='size-4' />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align='end' className='w-32'>
              <DropdownMenuItem onClick={() => onEdit(config)}>
                <Edit2 className='mr-2 size-3.5' />
                <span>{t('editCard')}</span>
              </DropdownMenuItem>
              <DropdownMenuItem
                className='text-destructive focus:text-destructive'
                onClick={() => onDelete(config)}
              >
                <Trash2 className='mr-2 size-3.5' />
                <span>{t('deleteCard')}</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>

      <CardContent className='space-y-3 px-5 py-2 text-sm'>
        {/* Venue and Seat Target */}
        <div className='grid grid-cols-2 gap-2 rounded-lg bg-muted/40 p-2.5 border border-border/30'>
          <div className='space-y-0.5 min-w-0'>
            <span className='text-[11px] text-muted-foreground font-medium flex items-center gap-1'>
              <Building2 className='size-3 text-muted-foreground' />
              {t('card.targetVenue')}
            </span>
            <p className='text-xs font-semibold truncate'>
              {config.library_name}{' '}
              <span className='text-muted-foreground font-normal'>
                ({config.floor}F)
              </span>
            </p>
          </div>

          <div className='space-y-0.5 min-w-0'>
            <span className='text-[11px] text-muted-foreground font-medium flex items-center gap-1'>
              <MapPin className='size-3 text-muted-foreground' />
              {t('card.targetSeat')}
            </span>
            <p className='text-xs font-semibold truncate'>
              {config.seat_name}{' '}
              <span className='text-muted-foreground font-mono font-normal text-[11px]'>
                [{config.seat_key}]
              </span>
            </p>
          </div>
        </div>

        {/* Status Badges */}
        <div className='flex items-center gap-2 flex-wrap pt-1'>
          {/* Auto Checkin status */}
          <Badge
            variant={config.auto_checkin ? 'default' : 'secondary'}
            className={cn(
              'text-[11px] gap-1 font-medium',
              config.auto_checkin
                ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-500/20 border-emerald-500/30'
                : 'bg-muted text-muted-foreground',
            )}
          >
            <Radio className='size-3' />
            {t('card.autoCheckin')}:{' '}
            {config.auto_checkin ? t('card.enabled') : t('card.disabled')}
          </Badge>

          {/* Session Credential Status */}
          <Badge
            variant='outline'
            className={cn(
              'text-[11px] gap-1 font-medium',
              isCookieValid
                ? 'text-emerald-600 dark:text-emerald-400 border-emerald-500/30'
                : 'text-amber-600 dark:text-amber-400 border-amber-500/30 bg-amber-500/10',
            )}
          >
            {isCookieValid ? (
              <ShieldCheck className='size-3' />
            ) : (
              <ShieldAlert className='size-3' />
            )}
            {isCookieValid ? t('card.authValid') : t('card.authExpired')}
          </Badge>

          {/* Checkin Credential Status if enabled */}
          {config.auto_checkin && (
            <Badge
              variant='outline'
              className={cn(
                'text-[11px] gap-1 font-medium',
                isCheckinValid
                  ? 'text-emerald-600 dark:text-emerald-400 border-emerald-500/30'
                  : 'text-amber-600 dark:text-amber-400 border-amber-500/30 bg-amber-500/10',
              )}
            >
              {isCheckinValid ? (
                <CheckCircle2 className='size-3' />
              ) : (
                <ShieldAlert className='size-3' />
              )}
              {isCheckinValid
                ? t('card.checkinValid')
                : t('card.checkinExpired')}
            </Badge>
          )}
        </div>
      </CardContent>

      <CardFooter className='pt-3 pb-5 px-5 border-t border-border/40 mt-3'>
        <Button
          onClick={() => onRun(config)}
          disabled={isRunning}
          className='w-full gap-2 font-medium shadow-sm transition-transform active:scale-[0.98]'
        >
          {isRunning ? (
            <>
              <Spinner className='size-4' />
              <span>{t('running')}</span>
            </>
          ) : (
            <>
              <Play className='size-4 fill-current' />
              <span>{t('runNow')}</span>
            </>
          )}
        </Button>
      </CardFooter>
    </Card>
  );
}
