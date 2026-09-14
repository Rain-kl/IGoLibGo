// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Save, Sliders } from 'lucide-react';
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
import { Switch } from '@/components/ui/switch';
import { Spinner } from '@/components/ui/spinner';
import { IGoService } from '@/lib/services/igo';
import type { SettingsResponse } from '@/lib/services/igo/types';

interface RuntimeSettingsTabProps {
  settings: SettingsResponse | null;
  loading: boolean;
  onRefresh: () => void;
}

export function RuntimeSettingsTab({
  settings,
  loading,
  onRefresh,
}: RuntimeSettingsTabProps) {
  const t = useTranslations('igo.settings');
  const tCommon = useTranslations('common');

  const [timeoutSeconds, setTimeoutSeconds] = React.useState(15);
  const [maxRetries, setMaxRetries] = React.useState(3);
  const [graphqlOverrides, setGraphqlOverrides] = React.useState(false);
  const [autoRelease, setAutoRelease] = React.useState(false);
  const [autoReleaseLead, setAutoReleaseLead] = React.useState(60);
  const [saving, setSaving] = React.useState(false);

  React.useEffect(() => {
    if (settings) {
      setTimeoutSeconds(settings.request_timeout_seconds || 15);
      setMaxRetries(settings.network_max_retries || 3);
      setGraphqlOverrides(settings.traceint_graphql_overrides_enabled || false);
      setAutoRelease(settings.auto_release_enabled || false);
      setAutoReleaseLead(settings.auto_release_lead_seconds || 60);
    }
  }, [settings]);

  const handleSave = async () => {
    setSaving(true);
    try {
      await IGoService.config.saveSettings({
        request_timeout_seconds: timeoutSeconds,
        network_max_retries: maxRetries,
        traceint_graphql_overrides_enabled: graphqlOverrides,
        auto_release_enabled: autoRelease,
        auto_release_lead_seconds: autoReleaseLead,
      });
      toast.success(tCommon('save'));
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-2'>
          <Sliders className='size-4 text-primary' />
          <CardTitle className='text-base font-semibold'>
            {t('tabRuntime')}
          </CardTitle>
        </div>
        <CardDescription>
          配置底层的网络重试、请求超时及自动释放防违规参数
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-6'>
        <div className='grid grid-cols-1 sm:grid-cols-2 gap-6'>
          <div className='space-y-2'>
            <Label htmlFor='timeout-sec' className='text-xs font-medium'>
              {t('timeoutSeconds')}
            </Label>
            <Input
              id='timeout-sec'
              type='number'
              value={timeoutSeconds}
              onChange={(e) =>
                setTimeoutSeconds(Math.max(3, Number(e.target.value)))
              }
              min={3}
              max={120}
              className='h-9 text-xs font-mono'
            />
            <p className='text-[10px] text-muted-foreground'>
              单个 HTTP / GraphQL 请求的最大等待响应秒数
            </p>
          </div>

          <div className='space-y-2'>
            <Label htmlFor='max-retries' className='text-xs font-medium'>
              {t('maxRetries')}
            </Label>
            <Input
              id='max-retries'
              type='number'
              value={maxRetries}
              onChange={(e) =>
                setMaxRetries(Math.max(0, Number(e.target.value)))
              }
              min={0}
              max={10}
              className='h-9 text-xs font-mono'
            />
            <p className='text-[10px] text-muted-foreground'>
              遭遇网络超时或断网时的自动指数退避重试次数
            </p>
          </div>
        </div>

        <div className='divide-y border rounded-lg bg-muted/20'>
          <div className='p-3.5 flex items-center justify-between gap-4'>
            <div className='space-y-0.5'>
              <Label className='text-xs font-medium'>
                {t('graphqlOverrides')}
              </Label>
              <p className='text-[11px] text-muted-foreground'>
                允许使用自定义协议模板覆写原生 TraceInt GraphQL 查询语句
              </p>
            </div>
            <Switch
              checked={graphqlOverrides}
              onCheckedChange={setGraphqlOverrides}
              aria-label={t('graphqlOverrides')}
            />
          </div>

          <div className='p-3.5 flex items-center justify-between gap-4'>
            <div className='space-y-0.5'>
              <Label className='text-xs font-medium'>{t('autoRelease')}</Label>
              <p className='text-[11px] text-muted-foreground'>
                在预约即将到期前自动释放，避免因离馆忘记退座造成违约记录
              </p>
            </div>
            <Switch
              checked={autoRelease}
              onCheckedChange={setAutoRelease}
              aria-label={t('autoRelease')}
            />
          </div>

          {autoRelease && (
            <div className='p-3.5 space-y-2 bg-muted/40'>
              <Label htmlFor='lead-sec' className='text-xs font-medium'>
                {t('autoReleaseLead')}
              </Label>
              <Input
                id='lead-sec'
                type='number'
                value={autoReleaseLead}
                onChange={(e) =>
                  setAutoReleaseLead(Math.max(10, Number(e.target.value)))
                }
                min={10}
                max={600}
                className='h-9 text-xs font-mono max-w-xs'
              />
            </div>
          )}
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
