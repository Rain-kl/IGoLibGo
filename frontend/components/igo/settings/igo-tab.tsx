// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Code2, Sliders } from 'lucide-react';

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { IGoService } from '@/lib/services/igo';
import type {
  ProtocolTemplatesResponse,
  SettingsResponse,
} from '@/lib/services/igo/types';

import { RuntimeSettingsTab } from './runtime-settings-tab';
import { ProtocolTemplatesTab } from './protocol-templates-tab';

export function IGoTab() {
  const t = useTranslations('igo.settings');

  const [settings, setSettings] = React.useState<SettingsResponse | null>(null);
  const [templates, setTemplates] =
    React.useState<ProtocolTemplatesResponse | null>(null);
  const [loading, setLoading] = React.useState(true);

  const loadData = React.useCallback(async () => {
    setLoading(true);
    try {
      const [settingsRes, templatesRes] = await Promise.all([
        IGoService.config.getSettings().catch(() => null),
        IGoService.config.getProtocolTemplates().catch(() => null),
      ]);
      setSettings(settingsRes);
      setTemplates(templatesRes);
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    loadData();
  }, [loadData]);

  return (
    <div className='space-y-4 w-full'>
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
        </TabsList>

        <TabsContent
          value='runtime'
          className='space-y-4 focus-visible:outline-none'
        >
          <RuntimeSettingsTab
            settings={settings}
            loading={loading}
            onRefresh={loadData}
          />
        </TabsContent>

        <TabsContent
          value='templates'
          className='space-y-4 focus-visible:outline-none'
        >
          <ProtocolTemplatesTab
            templates={templates}
            loading={loading}
            onRefresh={loadData}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
