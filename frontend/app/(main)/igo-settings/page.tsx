// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Bell, Cloud, Code2, FileText, Settings, Sliders } from 'lucide-react';

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { IGoService } from '@/lib/services/igo';
import type {
  ProtocolTemplatesResponse,
  SettingsResponse,
  WebDAVSettings,
} from '@/lib/services/igo/types';

import { RuntimeSettingsTab } from '@/components/igo/settings/runtime-settings-tab';
import { ProtocolTemplatesTab } from '@/components/igo/settings/protocol-templates-tab';
import { BackupRestoreTab } from '@/components/igo/settings/backup-restore-tab';
import { WebDAVSyncTab } from '@/components/igo/settings/webdav-sync-tab';
import { NotificationJumpCard } from '@/components/igo/settings/notification-jump-card';

export default function IGoSettingsPage() {
  const t = useTranslations('igo.settings');

  const [settings, setSettings] = React.useState<SettingsResponse | null>(null);
  const [templates, setTemplates] =
    React.useState<ProtocolTemplatesResponse | null>(null);
  const [webdav, setWebdav] = React.useState<WebDAVSettings | null>(null);
  const [loading, setLoading] = React.useState(true);

  const loadData = React.useCallback(async () => {
    setLoading(true);
    try {
      const [settingsRes, templatesRes, webdavRes] = await Promise.all([
        IGoService.config.getSettings().catch(() => null),
        IGoService.config.getProtocolTemplates().catch(() => null),
        IGoService.config.getWebDAV().catch(() => null),
      ]);
      setSettings(settingsRes);
      setTemplates(templatesRes);
      setWebdav(webdavRes);
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    loadData();
  }, [loadData]);

  return (
    <div className='py-6 px-1 space-y-6 w-full'>
      {/* 1. 标准标题 */}
      <div className='flex items-center gap-2'>
        <Settings className='size-5 text-primary' />
        <div>
          <h1 className='text-2xl font-semibold tracking-tight'>
            {t('title')}
          </h1>
          <p className='text-sm text-muted-foreground mt-0.5'>
            {t('description')}
          </p>
        </div>
      </div>

      {/* 2. 标签页切换 */}
      <Tabs defaultValue='runtime' className='space-y-4'>
        <TabsList className='bg-muted/60 p-1 flex-wrap h-auto gap-1'>
          <TabsTrigger value='runtime' className='text-xs gap-1.5'>
            <Sliders className='size-3.5' />
            <span>{t('tabRuntime')}</span>
          </TabsTrigger>
          <TabsTrigger value='templates' className='text-xs gap-1.5'>
            <Code2 className='size-3.5' />
            <span>{t('tabTemplates')}</span>
          </TabsTrigger>
          <TabsTrigger value='backup' className='text-xs gap-1.5'>
            <FileText className='size-3.5' />
            <span>{t('tabBackup')}</span>
          </TabsTrigger>
          <TabsTrigger value='webdav' className='text-xs gap-1.5'>
            <Cloud className='size-3.5' />
            <span>{t('tabWebdav')}</span>
          </TabsTrigger>
          <TabsTrigger value='notifications' className='text-xs gap-1.5'>
            <Bell className='size-3.5' />
            <span>{t('tabNotifications')}</span>
          </TabsTrigger>
        </TabsList>

        <TabsContent value='runtime' className='space-y-4'>
          <RuntimeSettingsTab
            settings={settings}
            loading={loading}
            onRefresh={loadData}
          />
        </TabsContent>

        <TabsContent value='templates' className='space-y-4'>
          <ProtocolTemplatesTab
            templates={templates}
            loading={loading}
            onRefresh={loadData}
          />
        </TabsContent>

        <TabsContent value='backup' className='space-y-4'>
          <BackupRestoreTab />
        </TabsContent>

        <TabsContent value='webdav' className='space-y-4'>
          <WebDAVSyncTab
            webdav={webdav}
            loading={loading}
            onRefresh={loadData}
          />
        </TabsContent>

        <TabsContent value='notifications' className='space-y-4'>
          <NotificationJumpCard />
        </TabsContent>
      </Tabs>
    </div>
  );
}
