// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Send } from 'lucide-react';
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { IGoService } from '@/lib/services/igo';
import { ApiErrorBase } from '@/lib/services/core/errors';
import type { AccountDTO, CheckInInfoDTO } from '@/lib/services/igo/types';

interface CheckInSignBarProps {
  accounts: AccountDTO[];
  infos: CheckInInfoDTO[];
  defaultLibraryId?: number;
  defaultLibraryName?: string;
  onNeedCheckinAuth: (accountId: string) => void;
}

export function CheckInSignBar({
  accounts,
  infos,
  defaultLibraryId,
  defaultLibraryName,
  onNeedCheckinAuth,
}: CheckInSignBarProps) {
  const t = useTranslations('igo.checkin');
  const tCommon = useTranslations('common');
  const [accountId, setAccountId] = React.useState('');
  const [infoId, setInfoId] = React.useState('');
  const [libraryId, setLibraryId] = React.useState(
    defaultLibraryId ? String(defaultLibraryId) : '',
  );
  const [signing, setSigning] = React.useState(false);

  React.useEffect(() => {
    if (defaultLibraryId) setLibraryId(String(defaultLibraryId));
  }, [defaultLibraryId]);

  const handleSign = async () => {
    const lib = Number(libraryId);
    if (!accountId || !infoId || !lib) {
      toast.error(t('signMissing'));
      return;
    }
    setSigning(true);
    try {
      const res = await IGoService.checkin.signInfo(infoId, {
        account_id: accountId,
        expected_library_id: lib,
        expected_library_name: defaultLibraryName,
      });
      toast.success(res.message || t('signSuccess'));
    } catch (err) {
      if (
        err instanceof ApiErrorBase &&
        (err.code === 'need_auth' || err.statusCode === 409)
      ) {
        toast.error(err.message);
        onNeedCheckinAuth(accountId);
        return;
      }
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSigning(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className='text-base'>{t('signTitle')}</CardTitle>
        <CardDescription>{t('signDesc')}</CardDescription>
      </CardHeader>
      <CardContent className='grid gap-3 sm:grid-cols-4'>
        <div className='space-y-1.5'>
          <Label>{t('selectAccount')}</Label>
          <Select value={accountId} onValueChange={setAccountId}>
            <SelectTrigger aria-label={t('selectAccount')}>
              <SelectValue placeholder={t('selectAccount')} />
            </SelectTrigger>
            <SelectContent>
              {accounts.map((a) => (
                <SelectItem key={a.id} value={a.id}>
                  {a.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className='space-y-1.5'>
          <Label>{t('selectInfo')}</Label>
          <Select value={infoId} onValueChange={setInfoId}>
            <SelectTrigger aria-label={t('selectInfo')}>
              <SelectValue placeholder={t('selectInfo')} />
            </SelectTrigger>
            <SelectContent>
              {infos.map((info) => (
                <SelectItem key={info.id} value={info.id}>
                  {info.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className='space-y-1.5'>
          <Label htmlFor='sign-lib'>{t('libraryId')}</Label>
          <Input
            id='sign-lib'
            value={libraryId}
            onChange={(e) => setLibraryId(e.target.value)}
          />
        </div>
        <div className='flex items-end'>
          <Button className='w-full' onClick={handleSign} disabled={signing}>
            <Send className='size-4' />
            {t('signNowBtn')}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
