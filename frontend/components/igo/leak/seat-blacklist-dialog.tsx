// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Ban } from 'lucide-react';
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
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { IGoService } from '@/lib/services/igo';
import type { LibrarySummary, SeatRef } from '@/lib/services/igo/types';

interface SeatBlacklistDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  libraries: LibrarySummary[];
  blacklist: Record<number, SeatRef[]>;
  onRefreshBlacklist: () => void;
}

export function SeatBlacklistDialog({
  open,
  onOpenChange,
  libraries,
  blacklist,
  onRefreshBlacklist,
}: SeatBlacklistDialogProps) {
  const t = useTranslations('igo.leak');
  const tCommon = useTranslations('common');

  const [selectedLibId, setSelectedLibId] = React.useState<string>('');
  const [textInput, setTextInput] = React.useState('');
  const [saving, setSaving] = React.useState(false);

  React.useEffect(() => {
    if (libraries.length > 0 && !selectedLibId) {
      setSelectedLibId(String(libraries[0].library_id));
    }
  }, [libraries, selectedLibId]);

  React.useEffect(() => {
    if (selectedLibId) {
      const items = blacklist[Number(selectedLibId)] || [];
      setTextInput(items.map((s) => s.seat_name || s.seat_key).join('\n'));
    }
  }, [selectedLibId, blacklist]);

  const handleSave = async () => {
    if (!selectedLibId) return;
    setSaving(true);
    try {
      const libId = Number(selectedLibId);
      const lines = textInput
        .split('\n')
        .map((l) => l.trim())
        .filter(Boolean);

      const seatRefs: SeatRef[] = lines.map((line) => ({
        seat_key: line,
        seat_name: line,
      }));

      const updated = { ...blacklist, [libId]: seatRefs };
      await IGoService.task.saveGlobalLeakBlacklist({ items: updated });
      toast.success(tCommon('save'));
      onRefreshBlacklist();
      onOpenChange(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Ban className='size-5 text-destructive' />
            {t('blacklistDialogTitle')}
          </DialogTitle>
          <DialogDescription>{t('blacklistDesc')}</DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-2'>
          <div className='space-y-1.5'>
            <Label className='text-xs font-medium'>选择场馆</Label>
            <Select value={selectedLibId} onValueChange={setSelectedLibId}>
              <SelectTrigger className='h-9 text-xs'>
                <SelectValue placeholder='请选择场馆' />
              </SelectTrigger>
              <SelectContent>
                {libraries.map((lib) => (
                  <SelectItem
                    key={lib.library_id}
                    value={String(lib.library_id)}
                    className='text-xs'
                  >
                    {lib.name} ({lib.floor})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='blacklist-input' className='text-xs font-medium'>
              黑名单座位列表
            </Label>
            <Textarea
              id='blacklist-input'
              value={textInput}
              onChange={(e) => setTextInput(e.target.value)}
              placeholder={t('blacklistPlaceholder')}
              rows={6}
              className='font-mono text-xs'
            />
            <p className='text-[10px] text-muted-foreground'>
              每行输入一个不希望被捡漏的座位编号或 Key
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant='outline'
            size='sm'
            onClick={() => onOpenChange(false)}
          >
            {tCommon('cancel')}
          </Button>
          <Button size='sm' onClick={handleSave} disabled={saving}>
            {tCommon('save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
