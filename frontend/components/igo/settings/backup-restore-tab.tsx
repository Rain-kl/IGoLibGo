// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Download, FileDown, FileUp, Upload } from 'lucide-react';
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
import { Textarea } from '@/components/ui/textarea';
import { Spinner } from '@/components/ui/spinner';
import { IGoService } from '@/lib/services/igo';

export function BackupRestoreTab() {
  const t = useTranslations('igo.settings');
  const tCommon = useTranslations('common');

  // 导出
  const [exportPassword, setExportPassword] = React.useState('');
  const [exporting, setExporting] = React.useState(false);

  // 导入
  const [importPassword, setImportPassword] = React.useState('');
  const [importContent, setImportContent] = React.useState('');
  const [importing, setImporting] = React.useState(false);

  const handleExport = async () => {
    if (exportPassword.length < 8) {
      toast.error('备份密码长度不能少于 8 位');
      return;
    }

    setExporting(true);
    try {
      const res = await IGoService.config.exportBackup({
        password: exportPassword,
      });
      // 触发浏览器下载
      const blob = new Blob([res.content], {
        type: 'application/octet-stream',
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = res.filename || `igo-backup-${Date.now()}.igobk`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      toast.success('备份导出成功并已开始下载');
      setExportPassword('');
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setExporting(false);
    }
  };

  const handleImport = async () => {
    if (!importPassword) {
      toast.error('请输入还原密码');
      return;
    }
    if (!importContent.trim()) {
      toast.error('请粘贴或上传备份文件内容');
      return;
    }

    setImporting(true);
    try {
      await IGoService.config.importBackup({
        password: importPassword,
        content: importContent.trim(),
      });
      toast.success('备份还原成功！配置已刷新');
      setImportPassword('');
      setImportContent('');
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setImporting(false);
    }
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      if (text) {
        setImportContent(text);
        toast.success(`已读取备份文件: ${file.name}`);
      }
    };
    reader.readAsText(file);
  };

  return (
    <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
      {/* 导出卡片 */}
      <Card className='border-border/60 shadow-sm'>
        <CardHeader className='pb-3'>
          <div className='flex items-center gap-2'>
            <FileDown className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('backupExportTitle')}
            </CardTitle>
          </div>
          <CardDescription>{t('backupExportDesc')}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-1.5'>
            <Label htmlFor='export-pwd' className='text-xs font-medium'>
              {t('passwordPrompt')}
            </Label>
            <Input
              id='export-pwd'
              type='password'
              value={exportPassword}
              onChange={(e) => setExportPassword(e.target.value)}
              placeholder='至少 8 位高强度密码'
              className='h-9 text-xs'
            />
          </div>

          <Button
            variant='default'
            size='sm'
            onClick={handleExport}
            disabled={exporting || exportPassword.length < 8}
            className='gap-1.5'
          >
            {exporting ? (
              <Spinner className='size-3.5' />
            ) : (
              <Download className='size-3.5' />
            )}
            <span>{t('exportBtn')}</span>
          </Button>
        </CardContent>
      </Card>

      {/* 导入卡片 */}
      <Card className='border-border/60 shadow-sm'>
        <CardHeader className='pb-3'>
          <div className='flex items-center gap-2'>
            <FileUp className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('backupImportTitle')}
            </CardTitle>
          </div>
          <CardDescription>{t('backupImportDesc')}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-1.5'>
            <Label htmlFor='import-pwd' className='text-xs font-medium'>
              解密还原密码
            </Label>
            <Input
              id='import-pwd'
              type='password'
              value={importPassword}
              onChange={(e) => setImportPassword(e.target.value)}
              placeholder='输入导出时设定的密码'
              className='h-9 text-xs'
            />
          </div>

          <div className='space-y-1.5'>
            <div className='flex items-center justify-between'>
              <Label className='text-xs font-medium'>备份文件数据</Label>
              <label className='cursor-pointer text-xs text-primary hover:underline'>
                <span>选择本地文件</span>
                <input
                  type='file'
                  accept='.igobk,.txt,.json'
                  onChange={handleFileUpload}
                  className='hidden'
                />
              </label>
            </div>
            <Textarea
              value={importContent}
              onChange={(e) => setImportContent(e.target.value)}
              placeholder='粘贴导出的加密备份 Base64 文本或直接选择本地文件'
              rows={3}
              className='font-mono text-xs'
            />
          </div>

          <Button
            variant='outline'
            size='sm'
            onClick={handleImport}
            disabled={importing || !importPassword || !importContent.trim()}
            className='gap-1.5'
          >
            {importing ? (
              <Spinner className='size-3.5' />
            ) : (
              <Upload className='size-3.5' />
            )}
            <span>{t('importBtn')}</span>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
