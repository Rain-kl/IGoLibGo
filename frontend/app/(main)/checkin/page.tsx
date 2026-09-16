// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { MapPin } from 'lucide-react';
import { toast } from 'sonner';

import { IGoService } from '@/lib/services/igo';
import type { AccountDTO, CheckInInfoDTO } from '@/lib/services/igo/types';
import { AccountList } from '@/components/igo/checkin/account-list';
import { CheckInInfoList } from '@/components/igo/checkin/checkin-info-list';
import { CheckInSignBar } from '@/components/igo/checkin/checkin-sign-bar';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

export default function CheckInPage() {
  const t = useTranslations('igo.checkin');
  const tCommon = useTranslations('common');
  const [accounts, setAccounts] = React.useState<AccountDTO[]>([]);
  const [infos, setInfos] = React.useState<CheckInInfoDTO[]>([]);
  const [libraryId, setLibraryId] = React.useState<number | undefined>();
  const [libraryName, setLibraryName] = React.useState<string | undefined>();
  const [loading, setLoading] = React.useState(true);
  const [editingInfoId, setEditingInfoId] = React.useState<string | null>(null);
  const [authOpen, setAuthOpen] = React.useState(false);
  const [authKind, setAuthKind] = React.useState<'login' | 'checkin'>(
    'checkin',
  );
  const [authAccount, setAuthAccount] = React.useState<AccountDTO | null>(null);
  const [authInput, setAuthInput] = React.useState('');
  const [authSaving, setAuthSaving] = React.useState(false);

  const loadData = React.useCallback(async () => {
    setLoading(true);
    try {
      const [accs, list, bound] = await Promise.all([
        IGoService.account.list().catch(() => []),
        IGoService.checkin.listInfos().catch(() => []),
        IGoService.venue.getBoundLibrary().catch(() => null),
      ]);
      setAccounts(accs || []);
      setInfos(list || []);
      if (bound?.library?.library_id) {
        setLibraryId(bound.library.library_id);
        setLibraryName(bound.library.name);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    loadData();
  }, [loadData]);

  const openAuth = (account: AccountDTO, kind: 'login' | 'checkin') => {
    setAuthAccount(account);
    setAuthKind(kind);
    setAuthInput('');
    setAuthOpen(true);
  };

  const submitAuth = async () => {
    if (!authAccount || !authInput.trim()) return;
    setAuthSaving(true);
    try {
      if (authKind === 'login') {
        await IGoService.account.login(authAccount.id, {
          code: authInput.trim(),
        });
      } else {
        const res = await IGoService.account.checkinAuth(authAccount.id, {
          code: authInput.trim(),
        });
        const uuid = res.device?.beacon_uuids?.[0];
        const editing = infos.find((i) => i.id === editingInfoId);
        if (uuid && editing && !editing.beacon_uuid) {
          await IGoService.checkin.updateInfo(editing.id, {
            name: editing.name,
            beacon_uuid: uuid,
            major: editing.major,
            minor: editing.minor,
            latitude: editing.latitude,
            longitude: editing.longitude,
          });
        }
      }
      toast.success(t('authSuccess'));
      setAuthOpen(false);
      loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setAuthSaving(false);
    }
  };

  return (
    <div className='py-6 px-1 space-y-6 w-full'>
      <div className='flex items-center gap-2'>
        <MapPin className='size-5 text-primary' />
        <div>
          <h1 className='text-2xl font-semibold tracking-tight'>
            {t('title')}
          </h1>
          <p className='text-sm text-muted-foreground mt-0.5'>
            {t('pageDesc')}
          </p>
        </div>
      </div>

      <AccountList
        accounts={accounts}
        loading={loading}
        onRefresh={loadData}
        onAuthorize={openAuth}
      />
      <CheckInInfoList
        infos={infos}
        loading={loading}
        editingId={editingInfoId}
        onEditingIdChange={setEditingInfoId}
        onRefresh={loadData}
      />
      <CheckInSignBar
        accounts={accounts}
        infos={infos}
        defaultLibraryId={libraryId}
        defaultLibraryName={libraryName}
        onNeedCheckinAuth={(id) => {
          const acc = accounts.find((a) => a.id === id);
          if (acc) openAuth(acc, 'checkin');
        }}
      />

      <Dialog open={authOpen} onOpenChange={setAuthOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {authKind === 'login' ? t('authLogin') : t('authCheckin')}
            </DialogTitle>
            <DialogDescription>{t('manualDialogDesc')}</DialogDescription>
          </DialogHeader>
          <div className='space-y-1.5'>
            <Label htmlFor='auth-code'>{t('qrLinkInputLabel')}</Label>
            <Textarea
              id='auth-code'
              value={authInput}
              onChange={(e) => setAuthInput(e.target.value)}
              placeholder={t('qrLinkPlaceholder')}
            />
          </div>
          <DialogFooter>
            <Button
              onClick={submitAuth}
              disabled={authSaving || !authInput.trim()}
            >
              {t('parseAndLoginBtn')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
