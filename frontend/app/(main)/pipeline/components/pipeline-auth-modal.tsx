// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';
import { ExternalLink, Play, ShieldAlert } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';
import type { PipelineConfigDTO } from '@/lib/services/igo/types';

const WECHAT_LOGIN_AUTH_URL =
  'https://open.weixin.qq.com/connect/oauth2/authorize?appid=wx2996d437cd442527&redirect_uri=https%3A%2F%2Fwechat.v2.traceint.com%2Findex.php%2FurlSign%2FgetUrl%3Furl%3Dhttps%253A%252F%252Fweb.traceint.com%252Fweb%252Findex.html&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect';

const WECHAT_CHECKIN_AUTH_URL =
  'https://open.weixin.qq.com/connect/oauth2/authorize?appid=wx2996d437cd442527&redirect_uri=https%3A%2F%2Fwechat.v2.traceint.com%2Findex.php%2FurlSign%2FgetUrl%3Furl%3Dhttps%253A%252F%252Fweb.traceint.com%252Fweb%252Findex.html%2523%252Fbeacon-checkin&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect';

interface PipelineAuthModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config: PipelineConfigDTO | null;
  needAuth: 'LOGIN' | 'CHECKIN' | null;
  onSubmitAndRun: (input: string) => Promise<void>;
  isLoading?: boolean;
}

export function PipelineAuthModal({
  open,
  onOpenChange,
  config,
  needAuth,
  onSubmitAndRun,
  isLoading,
}: PipelineAuthModalProps) {
  const t = useTranslations('igo.pipeline');
  const [authInput, setAuthInput] = React.useState('');

  React.useEffect(() => {
    if (open) {
      setAuthInput('');
    }
  }, [open]);

  if (!config || !needAuth) return null;

  const isLoginAuth = needAuth === 'LOGIN';
  const targetOAuthUrl = isLoginAuth
    ? WECHAT_LOGIN_AUTH_URL
    : WECHAT_CHECKIN_AUTH_URL;

  const handleSubmit = async () => {
    if (!authInput.trim()) {
      toast.error('请输入授权链接或 Code');
      return;
    }
    await onSubmitAndRun(authInput.trim());
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-md p-0 gap-0 overflow-hidden'>
        <DialogHeader className='px-6 pt-6 pb-4 border-b border-border/40'>
          <div className='flex items-center gap-2.5'>
            <div className='flex size-9 shrink-0 items-center justify-center rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400'>
              <ShieldAlert className='size-5' />
            </div>
            <div>
              <DialogTitle className='text-base font-semibold'>
                {t('authModal.title')}
              </DialogTitle>
              <DialogDescription className='text-xs mt-0.5'>
                配置「{config.name}」({config.id})
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className='px-6 py-5 space-y-4 text-sm'>
          <p className='text-xs text-muted-foreground'>{t('authModal.desc')}</p>

          <div className='rounded-lg bg-muted/40 border border-border/40 p-3.5 space-y-3'>
            <div className='flex items-center justify-between'>
              <span className='text-xs font-semibold'>
                {isLoginAuth
                  ? t('authModal.loginAuthTitle')
                  : t('authModal.checkinAuthTitle')}
              </span>
              <a
                href={targetOAuthUrl}
                target='_blank'
                rel='noreferrer'
                className='text-xs text-primary hover:underline flex items-center gap-1 font-medium'
              >
                <span>
                  {isLoginAuth
                    ? t('authModal.openLoginAuth')
                    : t('authModal.openCheckinAuth')}
                </span>
                <ExternalLink className='size-3' />
              </a>
            </div>

            <div className='space-y-1.5'>
              <Label
                htmlFor='reauth-input'
                className='text-xs text-muted-foreground'
              >
                {t('authModal.pastePlaceholder')}
              </Label>
              <Input
                id='reauth-input'
                value={authInput}
                onChange={(e) => setAuthInput(e.target.value)}
                placeholder='如 https://... 或 32 位 Code'
                className='text-xs font-mono'
              />
            </div>
          </div>
        </div>

        <DialogFooter className='px-6 py-4 border-t border-border/40 bg-muted/10 flex items-center justify-end gap-2'>
          <Button
            type='button'
            variant='outline'
            size='sm'
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
          >
            取消
          </Button>
          <Button
            type='button'
            size='sm'
            onClick={handleSubmit}
            disabled={isLoading || !authInput.trim()}
            className='gap-1.5'
          >
            {isLoading ? (
              <>
                <Spinner className='size-3.5' />
                <span>执行中...</span>
              </>
            ) : (
              <>
                <Play className='size-3.5 fill-current' />
                <span>{t('authModal.submitAndRun')}</span>
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
