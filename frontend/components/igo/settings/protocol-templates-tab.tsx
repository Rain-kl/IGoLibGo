// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Code2, RotateCcw, Save } from 'lucide-react';
import { toast } from 'sonner';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
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
import { Spinner } from '@/components/ui/spinner';
import { IGoService } from '@/lib/services/igo';
import type { ProtocolTemplatesResponse } from '@/lib/services/igo/types';

interface ProtocolTemplatesTabProps {
  templates: ProtocolTemplatesResponse | null;
  loading: boolean;
  onRefresh: () => void;
}

type TemplateField = keyof ProtocolTemplatesResponse;

const TEMPLATE_FIELDS: { key: TemplateField; label: string }[] = [
  { key: 'graphql_endpoint_url', label: 'GraphQL 端点 URL' },
  { key: 'reserve_seat_template', label: 'ReserveSeat (抢座/预约模板)' },
  { key: 'query_libraries_template', label: 'QueryLibraries (场馆列表模板)' },
  {
    key: 'query_library_layout_template',
    label: 'QueryLibraryLayout (座位布局模板)',
  },
  {
    key: 'query_library_rule_template',
    label: 'QueryLibraryRule (开放规则模板)',
  },
  {
    key: 'query_reservation_info_template',
    label: 'QueryReservationInfo (预约详情模板)',
  },
  {
    key: 'cancel_reservation_template',
    label: 'CancelReservation (取消预约模板)',
  },
  {
    key: 'tomorrow_reservation_warmup_template',
    label: 'TomorrowWarmup (明日预约预热模板)',
  },
  {
    key: 'tomorrow_reservation_save_template',
    label: 'TomorrowSave (明日预约提交模板)',
  },
];

export function ProtocolTemplatesTab({
  templates,
  loading,
  onRefresh,
}: ProtocolTemplatesTabProps) {
  const t = useTranslations('igo.settings');
  const tCommon = useTranslations('common');

  const [currentField, setCurrentField] = React.useState<TemplateField>(
    'reserve_seat_template',
  );
  const [formData, setFormData] = React.useState<
    Partial<ProtocolTemplatesResponse>
  >({});
  const [saving, setSaving] = React.useState(false);
  const [resetting, setResetting] = React.useState(false);

  React.useEffect(() => {
    if (templates) {
      setFormData(templates);
    }
  }, [templates]);

  const handleTextChange = (val: string) => {
    setFormData((prev) => ({
      ...prev,
      [currentField]: val,
    }));
  };

  const handleSave = async () => {
    if (!templates) return;
    setSaving(true);
    try {
      await IGoService.config.saveProtocolTemplates({
        overrides: { ...templates, ...formData } as ProtocolTemplatesResponse,
      });
      toast.success(tCommon('save'));
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSaving(false);
    }
  };

  const handleReset = async () => {
    setResetting(true);
    try {
      const defs = await IGoService.config.resetProtocolTemplates();
      setFormData(defs);
      toast.success('已恢复内置默认协议模板');
      onRefresh();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setResetting(false);
    }
  };

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <Code2 className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('tabTemplates')}
            </CardTitle>
          </div>
          <Button
            variant='outline'
            size='sm'
            onClick={handleReset}
            disabled={resetting || loading}
            className='h-7 text-xs gap-1'
          >
            <RotateCcw
              className={`size-3 ${resetting ? 'animate-spin' : ''}`}
            />
            <span>{t('resetTemplatesBtn')}</span>
          </Button>
        </div>
        <CardDescription>{t('templatesNotice')}</CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='space-y-1.5'>
          <Label className='text-xs font-medium'>选择查看/编辑模板项</Label>
          <Select
            value={currentField}
            onValueChange={(val) => setCurrentField(val as TemplateField)}
          >
            <SelectTrigger className='h-9 text-xs'>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {TEMPLATE_FIELDS.map((f) => (
                <SelectItem key={f.key} value={f.key} className='text-xs'>
                  {f.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className='space-y-1.5'>
          <Label className='text-xs font-medium'>模板代码内容</Label>
          <Textarea
            value={formData[currentField] || ''}
            onChange={(e) => handleTextChange(e.target.value)}
            rows={10}
            className='font-mono text-xs leading-relaxed bg-muted/20'
          />
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
            <span>{t('saveTemplatesBtn')}</span>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
