// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';
import { Plus, RefreshCw, Workflow } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { IGoService } from '@/lib/services/igo';
import type {
  PipelineConfigDTO,
  PipelineExecutionResult,
} from '@/lib/services/igo/types';
import { PipelineCard } from './components/pipeline-card';
import { PipelineDialog } from './components/pipeline-dialog';
import { PipelineAuthModal } from './components/pipeline-auth-modal';
import { PipelineResultDialog } from './components/pipeline-result-dialog';

export default function PipelinePage() {
  const t = useTranslations('igo.pipeline');
  const tCommon = useTranslations('common');

  const [configs, setConfigs] = React.useState<PipelineConfigDTO[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [runningId, setRunningId] = React.useState<string | null>(null);

  // Dialog states
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [editingConfig, setEditingConfig] =
    React.useState<PipelineConfigDTO | null>(null);

  // Delete state
  const [deletingConfig, setDeletingConfig] =
    React.useState<PipelineConfigDTO | null>(null);
  const [isDeleting, setIsDeleting] = React.useState(false);

  // Re-auth modal state
  const [authModalOpen, setAuthModalOpen] = React.useState(false);
  const [authModalConfig, setAuthModalConfig] =
    React.useState<PipelineConfigDTO | null>(null);
  const [authModalNeed, setAuthModalNeed] = React.useState<
    'LOGIN' | 'CHECKIN' | null
  >(null);

  // Result dialog state
  const [resultDialogOpen, setResultDialogOpen] = React.useState(false);
  const [executionResult, setExecutionResult] =
    React.useState<PipelineExecutionResult | null>(null);

  const loadConfigs = React.useCallback(async () => {
    setLoading(true);
    try {
      const data = await IGoService.pipeline.listConfigs();
      setConfigs(data || []);
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : '加载配置列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    loadConfigs();
  }, [loadConfigs]);

  const handleCreate = () => {
    setEditingConfig(null);
    setDialogOpen(true);
  };

  const handleEdit = (cfg: PipelineConfigDTO) => {
    setEditingConfig(cfg);
    setDialogOpen(true);
  };

  const handleDeletePrompt = (cfg: PipelineConfigDTO) => {
    setDeletingConfig(cfg);
  };

  const handleConfirmDelete = async () => {
    if (!deletingConfig) return;
    setIsDeleting(true);
    try {
      await IGoService.pipeline.deleteConfig(deletingConfig.id);
      toast.success('配置已成功删除');
      setDeletingConfig(null);
      await loadConfigs();
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : '删除配置失败');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleRun = async (
    cfg: PipelineConfigDTO,
    overrideAuthInput?: string,
  ) => {
    setRunningId(cfg.id);
    try {
      let runReq: Record<string, string> | undefined;
      if (overrideAuthInput) {
        if (overrideAuthInput.includes('code=')) {
          if (authModalNeed === 'CHECKIN') {
            runReq = { checkin_url: overrideAuthInput };
          } else {
            runReq = { auth_url: overrideAuthInput };
          }
        } else if (overrideAuthInput.length === 32) {
          if (authModalNeed === 'CHECKIN') {
            runReq = { checkin_code: overrideAuthInput };
          } else {
            runReq = { auth_code: overrideAuthInput };
          }
        } else {
          runReq = { cookie: overrideAuthInput };
        }
      }

      const res = await IGoService.pipeline.runConfig(cfg.id, runReq);

      // Check if re-authentication is required
      if (res.need_auth === 'LOGIN' || res.need_auth === 'CHECKIN') {
        setAuthModalConfig(cfg);
        setAuthModalNeed(res.need_auth);
        setAuthModalOpen(true);
        return;
      }

      // Close auth modal if open
      setAuthModalOpen(false);

      // Show result dialog
      setExecutionResult(res);
      setResultDialogOpen(true);

      // Refresh configs list to reflect updated session / timestamps
      await loadConfigs();
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : '执行一条龙任务异常');
    } finally {
      setRunningId(null);
    }
  };

  return (
    <div className='w-full py-6 px-1 space-y-6'>
      {/* Header */}
      <div className='flex flex-col sm:flex-row sm:items-center justify-between gap-4'>
        <div className='flex items-center gap-2'>
          <Workflow className='size-5 text-primary' />
          <div>
            <h1 className='text-2xl font-semibold tracking-tight'>
              {t('title')}
            </h1>
            <p className='text-xs text-muted-foreground mt-0.5'>
              {t('description')}
            </p>
          </div>
        </div>

        <div className='flex items-center gap-2'>
          <Button
            variant='outline'
            size='sm'
            onClick={loadConfigs}
            disabled={loading}
            aria-label='刷新配置列表'
            className='gap-1.5 text-xs'
          >
            <RefreshCw
              className={`size-3.5 ${loading ? 'animate-spin' : ''}`}
            />
            <span>刷新</span>
          </Button>
          <Button
            size='sm'
            onClick={handleCreate}
            className='gap-1.5 text-xs shadow-sm'
          >
            <Plus className='size-3.5' />
            <span>{t('createCard')}</span>
          </Button>
        </div>
      </div>

      {/* Main Content: Loading / Empty / Grid */}
      {loading && configs.length === 0 ? (
        <div className='flex flex-col items-center justify-center py-20 gap-3 text-muted-foreground'>
          <Spinner className='size-8' />
          <p className='text-xs'>正在加载自动化卡片...</p>
        </div>
      ) : configs.length === 0 ? (
        <div className='flex flex-col items-center justify-center py-16 px-4 rounded-xl border border-dashed border-border/80 bg-muted/10 text-center space-y-3'>
          <div className='flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary'>
            <Workflow className='size-6' />
          </div>
          <div className='space-y-1 max-w-sm'>
            <p className='text-sm font-semibold'>{t('emptyTitle')}</p>
            <p className='text-xs text-muted-foreground'>{t('emptyDesc')}</p>
          </div>
          <Button
            size='sm'
            onClick={handleCreate}
            className='gap-1.5 text-xs mt-2'
          >
            <Plus className='size-3.5' />
            <span>{t('createCard')}</span>
          </Button>
        </div>
      ) : (
        <div className='grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'>
          {configs.map((cfg) => (
            <PipelineCard
              key={cfg.id}
              config={cfg}
              onRun={(c) => handleRun(c)}
              onEdit={handleEdit}
              onDelete={handleDeletePrompt}
              isRunning={runningId === cfg.id}
            />
          ))}
        </div>
      )}

      {/* Multi-step Create/Edit Wizard Dialog */}
      <PipelineDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editingConfig={editingConfig}
        onSuccess={loadConfigs}
      />

      {/* Expired Auth Modal */}
      <PipelineAuthModal
        open={authModalOpen}
        onOpenChange={setAuthModalOpen}
        config={authModalConfig}
        needAuth={authModalNeed}
        onSubmitAndRun={async (input) => {
          if (authModalConfig) {
            await handleRun(authModalConfig, input);
          }
        }}
        isLoading={Boolean(runningId)}
      />

      {/* Execution Result Dialog */}
      <PipelineResultDialog
        open={resultDialogOpen}
        onOpenChange={setResultDialogOpen}
        result={executionResult}
      />

      {/* Delete Confirmation Alert Dialog */}
      <AlertDialog
        open={Boolean(deletingConfig)}
        onOpenChange={(open) => !open && setDeletingConfig(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('deleteConfirm.title')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('deleteConfirm.desc')}
              {deletingConfig && (
                <span className='block mt-2 font-semibold text-foreground'>
                  {deletingConfig.name} ({deletingConfig.id})
                </span>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>
              {tCommon('cancel')}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmDelete}
              disabled={isDeleting}
              className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            >
              {isDeleting ? '正在删除...' : tCommon('delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
