// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Tag, Trash2 } from 'lucide-react';
import { toast } from 'sonner';

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import type { SeatSnapshot } from '@/lib/services/igo/types';

interface SeatLabelDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedSeats: SeatSnapshot[];
  currentLabels: Record<string, string>;
  onSaveLabel: (text: string) => Promise<void>;
  onDeleteLabel: (seatKeys: string[]) => Promise<void>;
}

export function SeatLabelDialog({
  open,
  onOpenChange,
  selectedSeats,
  currentLabels,
  onSaveLabel,
  onDeleteLabel,
}: SeatLabelDialogProps) {
  const t = useTranslations('igo.venue');
  const tCommon = useTranslations('common');

  const [text, setText] = React.useState('');
  const [submitting, setSubmitting] = React.useState(false);

  React.useEffect(() => {
    if (selectedSeats.length === 1) {
      setText(currentLabels[selectedSeats[0].seat_key] || '');
    } else {
      setText('');
    }
  }, [selectedSeats, currentLabels]);

  const handleSave = async () => {
    if (!text.trim() || selectedSeats.length === 0) return;
    setSubmitting(true);
    try {
      await onSaveLabel(text.trim());
      toast.success(tCommon('save'));
      onOpenChange(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (selectedSeats.length === 0) return;
    setSubmitting(true);
    try {
      await onDeleteLabel(selectedSeats.map((s) => s.seat_key));
      toast.success(tCommon('delete'));
      onOpenChange(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Tag className='size-5 text-primary' />
            {t('labelDialogTitle')}
          </DialogTitle>
          <DialogDescription>
            为当前选中的座位添加个性化备注标签（如“静音靠窗”、“考研专座”等）
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-2'>
          <div className='space-y-1.5'>
            <Label className='text-xs text-muted-foreground'>
              当前选中的目标座位 ({selectedSeats.length} 个)
            </Label>
            {selectedSeats.length === 0 ? (
              <div className='p-2 rounded bg-muted/30 text-xs text-muted-foreground'>
                未在座位图中选中座位，请先在座位图上勾选座位
              </div>
            ) : (
              <div className='flex flex-wrap gap-1 max-h-24 overflow-y-auto p-1.5 rounded border bg-muted/20'>
                {selectedSeats.map((s) => (
                  <Badge
                    key={s.seat_key}
                    variant='secondary'
                    className='text-[10px] font-mono'
                  >
                    {s.seat_name || s.seat_key}
                  </Badge>
                ))}
              </div>
            )}
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='label-input' className='text-xs font-medium'>
              备注文本
            </Label>
            <Input
              id='label-input'
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder={t('labelInputPlaceholder')}
              className='h-9 text-xs'
            />
          </div>
        </div>

        <DialogFooter className='sm:justify-between'>
          <Button
            variant='ghost'
            size='sm'
            onClick={handleDelete}
            disabled={submitting || selectedSeats.length === 0}
            className='text-destructive hover:bg-destructive/10 gap-1'
          >
            <Trash2 className='size-3.5' />
            {t('deleteLabel')}
          </Button>

          <div className='flex items-center gap-2'>
            <Button
              variant='outline'
              size='sm'
              onClick={() => onOpenChange(false)}
            >
              {tCommon('cancel')}
            </Button>
            <Button
              size='sm'
              onClick={handleSave}
              disabled={
                submitting || !text.trim() || selectedSeats.length === 0
              }
            >
              {tCommon('save')}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
