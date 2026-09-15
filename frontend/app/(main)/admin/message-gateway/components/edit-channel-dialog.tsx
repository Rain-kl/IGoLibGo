// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Loader2 } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  TelegramForm,
  type TelegramFormValue,
} from '../channels/telegram/form';
import { QQForm, type QQFormValue } from '../channels/qq/form';
import type {
  MessageChannel,
  UpdateMessageChannelRequest,
} from '@/lib/services/message-gateway';

interface EditChannelDialogProps {
  channel: MessageChannel | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (id: string, data: UpdateMessageChannelRequest) => void;
  submitting?: boolean;
}

export function EditChannelDialog({
  channel,
  open,
  onOpenChange,
  onSubmit,
  submitting,
}: EditChannelDialogProps) {
  const t = useTranslations('admin.messageGateway');
  const [name, setName] = React.useState('');
  const [telegram, setTelegram] = React.useState<TelegramFormValue>({
    bot_token: '',
    base_url: '',
  });
  const [qq, setQQ] = React.useState<QQFormValue>({
    app_id: '',
    app_secret: '',
    portal_host: '',
  });

  React.useEffect(() => {
    if (channel && open) {
      setName(channel.name || '');
      if (channel.type === 'telegram') {
        setTelegram({
          bot_token: '',
          base_url: channel.credentials?.api_base || '',
        });
      } else if (channel.type === 'qq') {
        setQQ({
          app_id: channel.credentials?.app_id || '',
          app_secret: '',
          portal_host: channel.extra?.portal_host || '',
        });
      }
    }
  }, [channel, open]);

  const canSubmit =
    Boolean(channel) &&
    name.trim() !== '' &&
    (channel?.type !== 'qq' || qq.app_id.trim() !== '');

  const handleSubmit = () => {
    if (!canSubmit || !channel) return;
    const credentials: Record<string, string> = {};
    const extra: Record<string, string> = {};

    if (channel.type === 'telegram') {
      if (telegram.bot_token.trim()) {
        credentials.token = telegram.bot_token.trim();
      }
      if (telegram.base_url.trim()) {
        credentials.api_base = telegram.base_url.trim();
      }
    } else if (channel.type === 'qq') {
      credentials.app_id = qq.app_id.trim();
      if (qq.app_secret.trim()) {
        credentials.client_secret = qq.app_secret.trim();
      }
      if (qq.portal_host.trim()) {
        extra.portal_host = qq.portal_host.trim();
      }
    }

    const data: UpdateMessageChannelRequest = {
      name: name.trim(),
      ...(Object.keys(credentials).length > 0 ? { credentials } : {}),
      ...(Object.keys(extra).length > 0 ? { extra } : {}),
    };
    onSubmit(channel.id, data);
  };

  if (!channel) return null;

  const typeLabel =
    channel.type === 'qq'
      ? t('typeQQ')
      : channel.type === 'telegram'
        ? t('typeTelegram')
        : channel.type;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('editChannelTitle')}</DialogTitle>
        </DialogHeader>
        <div className='space-y-4'>
          <div className='space-y-2'>
            <Label htmlFor='mg-edit-name'>{t('name')}</Label>
            <Input
              id='mg-edit-name'
              value={name}
              placeholder={t('namePlaceholder')}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className='space-y-2'>
            <Label>{t('type')}</Label>
            <Input
              value={typeLabel}
              disabled
              className='bg-muted cursor-not-allowed'
            />
          </div>
          {channel.type === 'telegram' ? (
            <TelegramForm
              value={telegram}
              onChange={setTelegram}
              keepSecretHint
            />
          ) : null}
          {channel.type === 'qq' ? (
            <QQForm value={qq} onChange={setQQ} keepSecretHint />
          ) : null}
        </div>
        <DialogFooter>
          <Button
            variant='outline'
            onClick={() => onOpenChange(false)}
            disabled={submitting}
          >
            {t('cancel')}
          </Button>
          <Button onClick={handleSubmit} disabled={!canSubmit || submitting}>
            {submitting ? <Loader2 className='size-4 animate-spin' /> : null}
            {t('save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
