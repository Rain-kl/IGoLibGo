// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Cloud, RefreshCw, Save, Send } from 'lucide-react';
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
import { Spinner } from '@/components/ui/spinner';
import { IGoService } from '@/lib/services/igo';
import type { WebDAVSettings } from '@/lib/services/igo/types';

interface WebDAVSyncTabProps {
  webdav: WebDAVSettings | null;
  loading: boolean;
  onRefresh: () => void;
}

export function WebDAVSyncTab({
  webdav,
  loading,
  onRefresh,
}: WebDAVSyncTabProps) {
  const t = useTranslations('igo.settings');
  const tCommon = useTranslations('common');

  const [endpoint, setEndpoint] = React.useState('');
  const [remoteDir, setRemoteDir] = React.useState('/IGoLibrary');
  const [username, setUsername] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [saving, setSaving] = React.useState(false);
  const [syncing, setSyncing] = React.useState(false);

  React.useEffect(() => {
    if (webdav) {
      setEndpoint(webdav.endpoint || '');
      setRemoteDir(webdav.remote_directory || '/IGoLibrary');
      setUsername(webdav.username || '');
    }
  }, [webdav]);

  const handleSave = async () => {
    if (!endpoint.trim()) {
      toast.error('请输入 WebDAV 端点地址');
      return;
    }

    setSaving(true);
    try {
      await IGoService.config.saveWebDAV({
        endpoint: endpoint.trim(),
        remote_directory: remoteDir.trim(),
        username: username.trim(),
        password: password ? password.trim() : undefined,
      });
      toast.success(tCommon('save'));
      setPassword('');
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSaving(false);
    }
  };

  const handleSync = async () => {
    setSyncing(true);
    try {
      const res = await IGoService.config.syncWebDAV();
      toast.success(res.message || 'WebDAV 双向同步完成');
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSyncing(false);
    }
  };

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <Cloud className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('webdavTitle')}
            </CardTitle>
          </div>
          <Button
            variant='outline'
            size='sm'
            onClick={handleSync}
            disabled={syncing || !webdav?.endpoint}
            className='h-7 text-xs gap-1'
          >
            <RefreshCw className={`size-3 ${syncing ? 'animate-spin' : ''}`} />
            <span>{t('testAndSyncBtn')}</span>
          </Button>
        </div>
        <CardDescription>{t('webdavDesc')}</CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
          <div className='space-y-1.5'>
            <Label htmlFor='webdav-url' className='text-xs font-medium'>
              {t('webdavEndpoint')}
            </Label>
            <Input
              id='webdav-url'
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder='https://dav.jianguoyun.com/dav/'
              className='h-9 text-xs font-mono'
            />
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='webdav-dir' className='text-xs font-medium'>
              {t('webdavRemoteDir')}
            </Label>
            <Input
              id='webdav-dir'
              value={remoteDir}
              onChange={(e) => setRemoteDir(e.target.value)}
              placeholder='/IGoLibrary'
              className='h-9 text-xs font-mono'
            />
          </div>
        </div>

        <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
          <div className='space-y-1.5'>
            <Label htmlFor='webdav-user' className='text-xs font-medium'>
              {t('webdavUsername')}
            </Label>
            <Input
              id='webdav-user'
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder='WebDAV 账号'
              className='h-9 text-xs'
            />
          </div>

          <div className='space-y-1.5'>
            <Label htmlFor='webdav-pwd' className='text-xs font-medium'>
              {t('webdavPassword')}
            </Label>
            <Input
              id='webdav-pwd'
              type='password'
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder={
                webdav?.password_set
                  ? t('webdavPasswordNotice')
                  : '应用授权独立密码'
              }
              className='h-9 text-xs'
            />
          </div>
        </div>

        <div className='pt-2'>
          <Button
            variant='default'
            size='sm'
            onClick={handleSave}
            disabled={saving || loading}
            className='gap-1.5'
          >
            {saving ? (
              <Spinner className='size-3.5' />
            ) : (
              <Save className='size-3.5' />
            )}
            <span>{tCommon('save')}</span>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
