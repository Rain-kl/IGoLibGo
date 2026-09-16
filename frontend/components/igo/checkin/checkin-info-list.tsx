// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Pencil, Plus, Radio, Trash2 } from 'lucide-react';
import { toast } from 'sonner';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { IGoService } from '@/lib/services/igo';
import type { CheckInInfoDTO } from '@/lib/services/igo/types';

interface CheckInInfoListProps {
  infos: CheckInInfoDTO[];
  loading: boolean;
  editingId: string | null;
  onEditingIdChange: (id: string | null) => void;
  onRefresh: () => void;
  onPrefillUUID?: (uuid: string) => void;
}

export function CheckInInfoList({
  infos,
  loading,
  editingId,
  onEditingIdChange,
  onRefresh,
}: CheckInInfoListProps) {
  const t = useTranslations('igo.checkin');
  const tCommon = useTranslations('common');
  const [open, setOpen] = React.useState(false);
  const [saving, setSaving] = React.useState(false);
  const [form, setForm] = React.useState({
    name: '',
    beacon_uuid: '',
    major: '10001',
    minor: '1980',
    latitude: '',
    longitude: '',
  });

  const startCreate = () => {
    onEditingIdChange(null);
    setForm({
      name: '',
      beacon_uuid: '',
      major: '10001',
      minor: '1980',
      latitude: '',
      longitude: '',
    });
    setOpen(true);
  };

  const startEdit = (info: CheckInInfoDTO) => {
    onEditingIdChange(info.id);
    setForm({
      name: info.name,
      beacon_uuid: info.beacon_uuid,
      major: String(info.major),
      minor: String(info.minor),
      latitude: info.latitude,
      longitude: info.longitude,
    });
    setOpen(true);
  };

  const handleSave = async () => {
    if (!form.name.trim()) return;
    setSaving(true);
    try {
      const payload = {
        name: form.name.trim(),
        beacon_uuid: form.beacon_uuid.trim(),
        major: Number(form.major) || 0,
        minor: Number(form.minor) || 0,
        latitude: form.latitude.trim(),
        longitude: form.longitude.trim(),
      };
      if (editingId) {
        await IGoService.checkin.updateInfo(editingId, payload);
      } else {
        await IGoService.checkin.createInfo(payload);
      }
      toast.success(t('infoSaved'));
      setOpen(false);
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await IGoService.checkin.deleteInfo(id);
      toast.success(t('infoDeleted'));
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    }
  };

  return (
    <Card>
      <CardHeader className='flex flex-row items-start justify-between gap-2'>
        <div>
          <CardTitle className='text-base'>{t('infosTitle')}</CardTitle>
          <CardDescription>{t('infosDesc')}</CardDescription>
        </div>
        <Button size='sm' onClick={startCreate} aria-label={t('createInfo')}>
          <Plus className='size-4' />
          {t('createInfo')}
        </Button>
      </CardHeader>
      <CardContent className='space-y-3'>
        {loading && infos.length === 0 ? (
          <p className='text-sm text-muted-foreground'>{tCommon('loading')}</p>
        ) : infos.length === 0 ? (
          <p className='text-sm text-muted-foreground'>{t('infosEmpty')}</p>
        ) : (
          infos.map((info) => (
            <div
              key={info.id}
              className='flex flex-col gap-2 rounded-md border p-3 sm:flex-row sm:items-center sm:justify-between'
            >
              <div className='space-y-1'>
                <div className='flex items-center gap-2'>
                  <Radio className='size-4 text-primary' />
                  <p className='font-medium'>{info.name}</p>
                </div>
                <p className='text-xs text-muted-foreground break-all'>
                  {info.beacon_uuid || t('beaconEmpty')}
                </p>
              </div>
              <div className='flex gap-2'>
                <Button
                  size='sm'
                  variant='outline'
                  onClick={() => startEdit(info)}
                  aria-label={t('editInfo')}
                >
                  <Pencil className='size-4' />
                </Button>
                <Button
                  size='sm'
                  variant='ghost'
                  onClick={() => handleDelete(info.id)}
                  aria-label={t('deleteInfo')}
                >
                  <Trash2 className='size-4' />
                </Button>
              </div>
            </div>
          ))
        )}
      </CardContent>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingId ? t('editInfo') : t('createInfo')}
            </DialogTitle>
            <DialogDescription>{t('infoFormDesc')}</DialogDescription>
          </DialogHeader>
          <div className='grid gap-3'>
            <div className='space-y-1.5'>
              <Label htmlFor='info-name'>{t('infoName')}</Label>
              <Input
                id='info-name'
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
              />
            </div>
            <div className='space-y-1.5'>
              <Label htmlFor='info-uuid'>UUID</Label>
              <Input
                id='info-uuid'
                value={form.beacon_uuid}
                onChange={(e) =>
                  setForm({ ...form, beacon_uuid: e.target.value })
                }
              />
            </div>
            <div className='grid grid-cols-2 gap-2'>
              <div className='space-y-1.5'>
                <Label htmlFor='info-major'>Major</Label>
                <Input
                  id='info-major'
                  value={form.major}
                  onChange={(e) => setForm({ ...form, major: e.target.value })}
                />
              </div>
              <div className='space-y-1.5'>
                <Label htmlFor='info-minor'>Minor</Label>
                <Input
                  id='info-minor'
                  value={form.minor}
                  onChange={(e) => setForm({ ...form, minor: e.target.value })}
                />
              </div>
            </div>
            <div className='grid grid-cols-2 gap-2'>
              <div className='space-y-1.5'>
                <Label htmlFor='info-lat'>{t('latitude')}</Label>
                <Input
                  id='info-lat'
                  value={form.latitude}
                  onChange={(e) =>
                    setForm({ ...form, latitude: e.target.value })
                  }
                />
              </div>
              <div className='space-y-1.5'>
                <Label htmlFor='info-lng'>{t('longitude')}</Label>
                <Input
                  id='info-lng'
                  value={form.longitude}
                  onChange={(e) =>
                    setForm({ ...form, longitude: e.target.value })
                  }
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={handleSave} disabled={saving || !form.name.trim()}>
              {tCommon('save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
