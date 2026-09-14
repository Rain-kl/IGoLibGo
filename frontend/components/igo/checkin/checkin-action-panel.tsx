// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { CheckCircle2, Compass, Send } from 'lucide-react';
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import { IGoService } from '@/lib/services/igo';
import type {
  BoundLibraryResponse,
  CheckInDeviceResponse,
  CheckInSignResponse,
} from '@/lib/services/igo/types';

interface CheckInActionPanelProps {
  boundInfo: BoundLibraryResponse | null;
  device: CheckInDeviceResponse | null;
  onSuccess: () => void;
}

export function CheckInActionPanel({
  boundInfo,
  device,
  onSuccess,
}: CheckInActionPanelProps) {
  const t = useTranslations('igo.checkin');
  const tCommon = useTranslations('common');

  const bound = boundInfo?.library;

  const [selectedBeacon, setSelectedBeacon] = React.useState<string>('');
  const [major, setMajor] = React.useState<number>(10001);
  const [minor, setMinor] = React.useState<number>(1980);
  const [latitude, setLatitude] = React.useState<number>(31.2304);
  const [longitude, setLongitude] = React.useState<number>(121.4737);

  const [signing, setSigning] = React.useState(false);
  const [signResult, setSignResult] =
    React.useState<CheckInSignResponse | null>(null);

  React.useEffect(() => {
    if (
      device?.beacon_uuids &&
      device.beacon_uuids.length > 0 &&
      !selectedBeacon
    ) {
      setSelectedBeacon(device.beacon_uuids[0]);
    }
  }, [device, selectedBeacon]);

  const handleSign = async () => {
    if (!bound) {
      toast.error('请先在【账户与场馆】中锁定场馆');
      return;
    }
    if (!selectedBeacon) {
      toast.error('请选择或输入 Beacon UUID');
      return;
    }

    setSigning(true);
    setSignResult(null);
    try {
      const res = await IGoService.checkin.signCheckIn({
        expected_library_id: bound.library_id,
        expected_library_name: bound.name,
        beacon_uuid: selectedBeacon,
        major,
        minor,
        latitude,
        longitude,
      });
      setSignResult(res);
      toast.success(res.message || t('signSuccess'));
      onSuccess();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSigning(false);
    }
  };

  return (
    <Card className='border-dashed shadow-none'>
      <CardHeader className='pb-3'>
        <div className='flex items-center gap-2'>
          <Compass className='size-4 text-primary' />
          <CardTitle className='text-base font-semibold'>
            {t('beaconTitle')}
          </CardTitle>
        </div>
        <CardDescription>
          设定打卡场馆基准 Beacon 与模拟地理坐标并执行远程签到
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
          {/* Beacon UUID */}
          <div className='space-y-1.5'>
            <Label className='text-xs font-medium'>{t('selectBeacon')}</Label>
            {device?.beacon_uuids && device.beacon_uuids.length > 0 ? (
              <Select value={selectedBeacon} onValueChange={setSelectedBeacon}>
                <SelectTrigger className='h-8 text-xs font-mono shadow-none bg-background'>
                  <SelectValue placeholder='请选择 Beacon UUID' />
                </SelectTrigger>
                <SelectContent>
                  {device.beacon_uuids.map((uuid) => (
                    <SelectItem
                      key={uuid}
                      value={uuid}
                      className='text-xs font-mono'
                    >
                      {uuid}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            ) : (
              <Input
                value={selectedBeacon}
                onChange={(e) => setSelectedBeacon(e.target.value)}
                placeholder='手动输入 Beacon UUID'
                className='h-8 text-xs font-mono shadow-none bg-background'
              />
            )}
          </div>

          {/* 场馆名称 */}
          <div className='space-y-1.5'>
            <Label className='text-xs font-medium'>当前目标场馆</Label>
            <Input
              value={bound ? `${bound.name} (${bound.floor})` : '未锁定场馆'}
              disabled
              className='h-8 text-xs bg-muted/30 shadow-none'
            />
          </div>
        </div>

        {/* 坐标与 Major/Minor */}
        <div className='grid grid-cols-2 sm:grid-cols-4 gap-3 pt-1'>
          <div className='space-y-1'>
            <Label className='text-[11px] text-muted-foreground'>Major</Label>
            <Input
              type='number'
              value={major}
              onChange={(e) => setMajor(Number(e.target.value))}
              className='h-8 text-xs font-mono shadow-none bg-background'
            />
          </div>
          <div className='space-y-1'>
            <Label className='text-[11px] text-muted-foreground'>Minor</Label>
            <Input
              type='number'
              value={minor}
              onChange={(e) => setMinor(Number(e.target.value))}
              className='h-8 text-xs font-mono shadow-none bg-background'
            />
          </div>
          <div className='space-y-1'>
            <Label className='text-[11px] text-muted-foreground'>
              纬度 (Latitude)
            </Label>
            <Input
              type='number'
              step='0.0001'
              value={latitude}
              onChange={(e) => setLatitude(Number(e.target.value))}
              className='h-8 text-xs font-mono shadow-none bg-background'
            />
          </div>
          <div className='space-y-1'>
            <Label className='text-[11px] text-muted-foreground'>
              经度 (Longitude)
            </Label>
            <Input
              type='number'
              step='0.0001'
              value={longitude}
              onChange={(e) => setLongitude(Number(e.target.value))}
              className='h-8 text-xs font-mono shadow-none bg-background'
            />
          </div>
        </div>

        {/* 打卡结果提示 */}
        {signResult && (
          <div className='p-3 rounded-lg border border-dashed border-emerald-500/30 bg-emerald-500/10 text-xs text-foreground flex items-start gap-2.5 shadow-none'>
            <CheckCircle2 className='size-4 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5' />
            <div className='space-y-0.5'>
              <div className='font-semibold text-emerald-900 dark:text-emerald-200'>
                {signResult.message || t('signSuccess')}
              </div>
              <div className='text-[11px] text-muted-foreground font-mono'>
                {signResult.signed_at && `签到时间: ${signResult.signed_at} `}
                {signResult.seat_name && `· 座位: ${signResult.seat_name} `}
                {signResult.expiration_time &&
                  `· 预约到期: ${signResult.expiration_time}`}
              </div>
            </div>
          </div>
        )}

        <div className='pt-2'>
          <Button
            variant='default'
            onClick={handleSign}
            disabled={signing || !bound || !selectedBeacon}
            className='gap-1.5 shadow-none'
          >
            {signing ? (
              <Spinner className='size-4' />
            ) : (
              <Send className='size-4' />
            )}
            <span>{signing ? t('signing') : t('signNowBtn')}</span>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
