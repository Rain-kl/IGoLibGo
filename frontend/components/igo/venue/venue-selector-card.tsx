// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import {
  Building2,
  CheckCircle2,
  Clock,
  Layers,
  Lock,
  RefreshCw,
  Users,
} from 'lucide-react';
import { toast } from 'sonner';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import { IGoService } from '@/lib/services/igo';
import type {
  BoundLibraryResponse,
  LibrarySummary,
} from '@/lib/services/igo/types';

interface VenueSelectorCardProps {
  boundInfo: BoundLibraryResponse | null;
  libraries: LibrarySummary[];
  loading: boolean;
  onRefresh: () => void;
  onSelectLibrary: (libId: number) => void;
}

export function VenueSelectorCard({
  boundInfo,
  libraries,
  loading,
  onRefresh,
  onSelectLibrary,
}: VenueSelectorCardProps) {
  const t = useTranslations('igo.venue');
  const tCommon = useTranslations('common');

  const [selectedId, setSelectedId] = React.useState<string>('');
  const [binding, setBinding] = React.useState(false);
  const [refreshing, setRefreshing] = React.useState(false);

  React.useEffect(() => {
    if (boundInfo?.library?.library_id) {
      setSelectedId(String(boundInfo.library.library_id));
    }
  }, [boundInfo]);

  const handleBind = async () => {
    if (!selectedId) return;
    setBinding(true);
    try {
      await IGoService.venue.bindLibrary(Number(selectedId));
      toast.success(tCommon('save'));
      onRefresh();
      onSelectLibrary(Number(selectedId));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setBinding(false);
    }
  };

  const handleRefreshBound = async () => {
    setRefreshing(true);
    try {
      await IGoService.venue.refreshBoundLibrary();
      toast.success(tCommon('save'));
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setRefreshing(false);
    }
  };

  const bound = boundInfo?.library;

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <Building2 className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('boundVenueTitle')}
            </CardTitle>
          </div>
          {loading ? (
            <Spinner className='size-4' />
          ) : bound ? (
            <Badge
              variant='outline'
              className='gap-1 border-primary/40 text-primary'
            >
              <CheckCircle2 className='size-3' />
              {t('currentBound')}
            </Badge>
          ) : (
            <Badge variant='secondary'>{t('noBound')}</Badge>
          )}
        </div>
        <CardDescription>
          {bound ? `${bound.name} · ${bound.floor}` : t('noBound')}
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        {bound ? (
          <div className='grid grid-cols-2 sm:grid-cols-4 gap-3 p-3 rounded-lg bg-muted/40 text-xs'>
            <div className='flex items-center gap-2'>
              <Building2 className='size-3.5 text-muted-foreground' />
              <div>
                <div className='text-muted-foreground'>场馆名称</div>
                <div className='font-medium truncate'>{bound.name}</div>
              </div>
            </div>
            <div className='flex items-center gap-2'>
              <Layers className='size-3.5 text-muted-foreground' />
              <div>
                <div className='text-muted-foreground'>所在楼层</div>
                <div className='font-medium'>{bound.floor}</div>
              </div>
            </div>
            <div className='flex items-center gap-2'>
              <Users className='size-3.5 text-muted-foreground' />
              <div>
                <div className='text-muted-foreground'>
                  {t('availableSeats')} / {t('totalSeats')}
                </div>
                <div className='font-semibold text-primary'>
                  {bound.total_seats - bound.used_seats - bound.booked_seats} /{' '}
                  {bound.total_seats}
                </div>
              </div>
            </div>
            <div className='flex items-center gap-2'>
              <Clock className='size-3.5 text-muted-foreground' />
              <div>
                <div className='text-muted-foreground'>场馆状态</div>
                <div className='font-medium'>
                  {bound.is_open ? (
                    <span className='text-emerald-500 font-medium'>开放中</span>
                  ) : (
                    <span className='text-muted-foreground'>已闭馆</span>
                  )}
                </div>
              </div>
            </div>
          </div>
        ) : null}

        <div className='flex flex-col sm:flex-row items-stretch sm:items-center gap-2 pt-1'>
          <div className='flex-1 min-w-48'>
            <Select
              value={selectedId}
              onValueChange={(val) => {
                setSelectedId(val);
                onSelectLibrary(Number(val));
              }}
            >
              <SelectTrigger className='h-9 text-xs'>
                <SelectValue placeholder={t('changeVenue')} />
              </SelectTrigger>
              <SelectContent>
                {libraries.map((lib) => (
                  <SelectItem
                    key={lib.library_id}
                    value={String(lib.library_id)}
                    className='text-xs'
                  >
                    {lib.name} ({lib.floor}) · 余{' '}
                    {lib.total_seats - lib.used_seats - lib.booked_seats}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className='flex items-center gap-2'>
            <Button
              variant='default'
              size='sm'
              onClick={handleBind}
              disabled={!selectedId || binding}
              className='gap-1.5'
            >
              {binding ? (
                <Spinner className='size-3.5' />
              ) : (
                <Lock className='size-3.5' />
              )}
              {t('bindVenueBtn')}
            </Button>
            {bound && (
              <Button
                variant='outline'
                size='sm'
                onClick={handleRefreshBound}
                disabled={refreshing}
                className='gap-1.5'
              >
                <RefreshCw
                  className={`size-3.5 ${refreshing ? 'animate-spin' : ''}`}
                />
                <span className='hidden sm:inline'>{t('refreshVenue')}</span>
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
