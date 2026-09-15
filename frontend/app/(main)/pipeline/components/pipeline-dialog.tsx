// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import type { PipelineConfigDTO } from '@/lib/services/igo/types';
import { PipelineCreateDialog } from './pipeline-create-dialog';
import { PipelineEditDialog } from './pipeline-edit-dialog';

export interface PipelineDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editingConfig?: PipelineConfigDTO | null;
  onSuccess: () => void;
}

export function PipelineDialog({
  open,
  onOpenChange,
  editingConfig,
  onSuccess,
}: PipelineDialogProps) {
  if (editingConfig) {
    return (
      <PipelineEditDialog
        open={open}
        onOpenChange={onOpenChange}
        config={editingConfig}
        onSuccess={onSuccess}
      />
    );
  }

  return (
    <PipelineCreateDialog
      open={open}
      onOpenChange={onOpenChange}
      onSuccess={onSuccess}
    />
  );
}

export { PipelineCreateDialog } from './pipeline-create-dialog';
export { PipelineEditDialog } from './pipeline-edit-dialog';
