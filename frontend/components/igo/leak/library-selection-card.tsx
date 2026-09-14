// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Building2, CheckSquare, Layers, Square } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import type {
  GlobalLeakLibraryTarget,
  LibrarySummary,
} from '@/lib/services/igo/types';

interface LibrarySelectionCardProps {
  allLibraries: LibrarySummary[];
  selectedLibraries: GlobalLeakLibraryTarget[];
  onSelectionChange: (selected: GlobalLeakLibraryTarget[]) => void;
}

export function LibrarySelectionCard({
  allLibraries,
  selectedLibraries,
  onSelectionChange,
}: LibrarySelectionCardProps) {
  const t = useTranslations('igo.leak');

  const selectedIds = React.useMemo(() => {
    return new Set(selectedLibraries.map((lib) => lib.library_id));
  }, [selectedLibraries]);

  const handleToggle = (lib: LibrarySummary) => {
    if (selectedIds.has(lib.library_id)) {
      onSelectionChange(
        selectedLibraries.filter((item) => item.library_id !== lib.library_id),
      );
    } else {
      onSelectionChange([
        ...selectedLibraries,
        {
          library_id: lib.library_id,
          library_name: lib.name,
          floor: lib.floor,
        },
      ]);
    }
  };

  const handleSelectAll = () => {
    onSelectionChange(
      allLibraries.map((lib) => ({
        library_id: lib.library_id,
        library_name: lib.name,
        floor: lib.floor,
      })),
    );
  };

  const handleClearAll = () => {
    onSelectionChange([]);
  };

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex flex-wrap items-center justify-between gap-2'>
          <div className='flex items-center gap-2'>
            <Building2 className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('venueScopeTitle')}
            </CardTitle>
            <Badge variant='outline' className='text-xs font-normal'>
              已选 {selectedLibraries.length} / 共 {allLibraries.length} 个场馆
            </Badge>
          </div>

          <div className='flex items-center gap-1.5'>
            <Button
              variant='outline'
              size='sm'
              onClick={handleSelectAll}
              className='h-7 text-xs gap-1'
            >
              <CheckSquare className='size-3' />
              <span>{t('selectAll')}</span>
            </Button>
            <Button
              variant='ghost'
              size='sm'
              onClick={handleClearAll}
              className='h-7 text-xs text-muted-foreground'
            >
              <Square className='size-3 mr-1' />
              <span>{t('clearAll')}</span>
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className='grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-2.5 max-h-96 overflow-y-auto pr-1'>
          {allLibraries.map((lib) => {
            const isChecked = selectedIds.has(lib.library_id);
            const availableSeats =
              lib.total_seats - lib.used_seats - lib.booked_seats;

            return (
              <div
                key={lib.library_id}
                onClick={() => handleToggle(lib)}
                className={`
                  p-2.5 rounded-lg border text-xs cursor-pointer transition-all flex items-start gap-2.5 select-none
                  ${
                    isChecked
                      ? 'border-primary/50 bg-primary/5 ring-1 ring-primary/20'
                      : 'border-border/60 bg-card hover:bg-muted/30'
                  }
                `}
              >
                <Checkbox
                  checked={isChecked}
                  onCheckedChange={() => handleToggle(lib)}
                  className='mt-0.5'
                />
                <div className='flex-1 min-w-0 space-y-0.5'>
                  <div className='font-semibold text-foreground truncate'>
                    {lib.name}
                  </div>
                  <div className='flex items-center gap-2 text-[10px] text-muted-foreground'>
                    <span className='flex items-center gap-0.5'>
                      <Layers className='size-2.5' />
                      {lib.floor}
                    </span>
                    <span>·</span>
                    <span className='text-primary font-medium'>
                      余 {availableSeats} 座
                    </span>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
