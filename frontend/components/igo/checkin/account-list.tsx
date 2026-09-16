// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { KeyRound, Plus, Trash2, UserRound } from 'lucide-react';
import { toast } from 'sonner';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
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
import type { AccountDTO } from '@/lib/services/igo/types';

interface AccountListProps {
  accounts: AccountDTO[];
  loading: boolean;
  onRefresh: () => void;
  onAuthorize: (account: AccountDTO, kind: 'login' | 'checkin') => void;
}

export function AccountList({
  accounts,
  loading,
  onRefresh,
  onAuthorize,
}: AccountListProps) {
  const t = useTranslations('igo.checkin');
  const tCommon = useTranslations('common');
  const [createOpen, setCreateOpen] = React.useState(false);
  const [name, setName] = React.useState('');
  const [saving, setSaving] = React.useState(false);

  const handleCreate = async () => {
    const trimmed = name.trim();
    if (!trimmed) return;
    setSaving(true);
    try {
      await IGoService.account.create({ name: trimmed });
      toast.success(t('accountCreated'));
      setCreateOpen(false);
      setName('');
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await IGoService.account.remove(id);
      toast.success(t('accountDeleted'));
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    }
  };

  return (
    <Card>
      <CardHeader className='flex flex-row items-start justify-between gap-2'>
        <div>
          <CardTitle className='text-base'>{t('accountsTitle')}</CardTitle>
          <CardDescription>{t('accountsDesc')}</CardDescription>
        </div>
        <Button
          size='sm'
          onClick={() => setCreateOpen(true)}
          aria-label={t('createAccount')}
        >
          <Plus className='size-4' />
          {t('createAccount')}
        </Button>
      </CardHeader>
      <CardContent className='space-y-3'>
        {loading && accounts.length === 0 ? (
          <p className='text-sm text-muted-foreground'>{tCommon('loading')}</p>
        ) : accounts.length === 0 ? (
          <p className='text-sm text-muted-foreground'>{t('accountsEmpty')}</p>
        ) : (
          accounts.map((acc) => (
            <div
              key={acc.id}
              className='flex flex-col gap-2 rounded-md border p-3 sm:flex-row sm:items-center sm:justify-between'
            >
              <div className='space-y-1'>
                <div className='flex items-center gap-2'>
                  <UserRound className='size-4 text-primary' />
                  <p className='font-medium'>{acc.name}</p>
                </div>
                <p className='text-xs text-muted-foreground'>
                  {acc.student_name || acc.nickname || t('noProfile')}
                </p>
                <div className='flex flex-wrap gap-1'>
                  <Badge variant={acc.has_cookie ? 'default' : 'secondary'}>
                    {acc.has_cookie ? t('cookieValid') : t('cookieMissing')}
                  </Badge>
                  <Badge
                    variant={acc.has_checkin_token ? 'default' : 'secondary'}
                  >
                    {acc.has_checkin_token
                      ? t('checkinValid')
                      : t('checkinMissing')}
                  </Badge>
                </div>
              </div>
              <div className='flex flex-wrap gap-2'>
                <Button
                  size='sm'
                  variant='outline'
                  onClick={() => onAuthorize(acc, 'login')}
                >
                  <KeyRound className='size-4' />
                  {t('authLogin')}
                </Button>
                <Button
                  size='sm'
                  variant='outline'
                  onClick={() => onAuthorize(acc, 'checkin')}
                >
                  {t('authCheckin')}
                </Button>
                <Button
                  size='sm'
                  variant='ghost'
                  aria-label={t('deleteAccount')}
                  onClick={() => handleDelete(acc.id)}
                >
                  <Trash2 className='size-4' />
                </Button>
              </div>
            </div>
          ))
        )}
      </CardContent>
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('createAccount')}</DialogTitle>
            <DialogDescription>{t('createAccountDesc')}</DialogDescription>
          </DialogHeader>
          <div className='space-y-1.5'>
            <Label htmlFor='account-name'>{t('accountName')}</Label>
            <Input
              id='account-name'
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button onClick={handleCreate} disabled={saving || !name.trim()}>
              {tCommon('save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
