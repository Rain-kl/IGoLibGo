// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Bookmark, Plus, Trash2, X } from 'lucide-react';
import { toast } from 'sonner';

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import type { SeatRef, SeatSnapshot } from '@/lib/services/igo/types';

interface FavoritesDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  favorites: SeatRef[];
  allSeats: SeatSnapshot[];
  onSelectSeat: (seat: SeatSnapshot) => void;
  onRemoveFavorite: (seatKey: string) => void;
}

export function FavoritesDrawer({
  open,
  onOpenChange,
  favorites,
  allSeats,
  onSelectSeat,
  onRemoveFavorite,
}: FavoritesDrawerProps) {
  const t = useTranslations('igo.venue');
  const tCommon = useTranslations('common');

  const seatMap = React.useMemo(() => {
    return new Map(allSeats.map((s) => [s.seat_key, s]));
  }, [allSeats]);

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className='sm:max-w-md flex flex-col'>
        <SheetHeader>
          <SheetTitle className='flex items-center gap-2 text-base'>
            <Bookmark className='size-4 text-amber-500 fill-amber-500' />
            {t('favoritesTitle')}
          </SheetTitle>
          <SheetDescription className='text-xs'>
            点击收藏的常用座位即可快速加入当前目标候选队列
          </SheetDescription>
        </SheetHeader>

        <div className='flex-1 overflow-hidden py-4'>
          {favorites.length === 0 ? (
            <div className='h-64 flex flex-col items-center justify-center text-center p-6 border-dashed border rounded-lg text-xs text-muted-foreground gap-2'>
              <Bookmark className='size-8 text-muted-foreground/30' />
              <span>{t('emptyFavorites')}</span>
            </div>
          ) : (
            <ScrollArea className='h-full pr-3'>
              <div className='space-y-2'>
                {favorites.map((fav, index) => {
                  const liveSeat = seatMap.get(fav.seat_key);
                  const isOccupied = liveSeat?.is_occupied ?? false;

                  return (
                    <div
                      key={fav.seat_key}
                      className='flex items-center justify-between p-2.5 rounded-lg border bg-card hover:bg-muted/40 transition-colors text-xs'
                    >
                      <div className='flex items-center gap-2'>
                        <span className='size-5 rounded bg-muted flex items-center justify-center font-mono font-medium text-[10px] text-muted-foreground'>
                          {index + 1}
                        </span>
                        <div>
                          <div className='font-semibold font-mono text-foreground'>
                            {fav.seat_name || fav.seat_key}
                          </div>
                          <div className='text-[10px] text-muted-foreground font-mono'>
                            Key: {fav.seat_key}
                          </div>
                        </div>
                      </div>

                      <div className='flex items-center gap-1.5'>
                        {isOccupied ? (
                          <Badge variant='secondary' className='text-[10px]'>
                            已被占用
                          </Badge>
                        ) : (
                          <Badge
                            variant='outline'
                            className='text-[10px] border-emerald-500/40 text-emerald-600 dark:text-emerald-400'
                          >
                            空闲
                          </Badge>
                        )}

                        {liveSeat && !isOccupied && (
                          <Button
                            variant='outline'
                            size='icon'
                            aria-label='添加此座位到候选'
                            onClick={() => {
                              onSelectSeat(liveSeat);
                              toast.success(
                                `已添加座位 ${fav.seat_name || fav.seat_key}`,
                              );
                            }}
                            className='size-7'
                          >
                            <Plus className='size-3.5' />
                          </Button>
                        )}

                        <Button
                          variant='ghost'
                          size='icon'
                          aria-label='从收藏中移除'
                          onClick={() => onRemoveFavorite(fav.seat_key)}
                          className='size-7 text-destructive hover:bg-destructive/10'
                        >
                          <Trash2 className='size-3.5' />
                        </Button>
                      </div>
                    </div>
                  );
                })}
              </div>
            </ScrollArea>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}
