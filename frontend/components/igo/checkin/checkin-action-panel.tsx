// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { AlertTriangle, CheckCircle2, Compass, Save, Send } from 'lucide-react';
import { toast } from 'sonner';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
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
  CheckInVenueProfile,
  ReservationResponse,
} from '@/lib/services/igo/types';

interface CheckInActionPanelProps {
  boundInfo: BoundLibraryResponse | null;
  device: CheckInDeviceResponse | null;
  currentReservation?: ReservationResponse | null;
  onSuccess: () => void;
}

export function CheckInActionPanel({
  boundInfo,
  device,
  currentReservation,
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

  const [customMode, setCustomMode] = React.useState<boolean>(false);
  const [savedProfile, setSavedProfile] =
    React.useState<CheckInVenueProfile | null>(null);
  const [isSaved, setIsSaved] = React.useState<boolean>(false);
  const [savingProfile, setSavingProfile] = React.useState<boolean>(false);

  const [signing, setSigning] = React.useState(false);
  const [signResult, setSignResult] =
    React.useState<CheckInSignResponse | null>(null);

  const checkIfMatchesSaved = React.useCallback(
    (
      b: string,
      maj: number,
      min: number,
      lat: number,
      lng: number,
      target?: CheckInVenueProfile | null,
    ) => {
      const p = target !== undefined ? target : savedProfile;
      if (!p) return false;
      return (
        p.beacon_uuid.trim().toUpperCase() === b.trim().toUpperCase() &&
        p.major === maj &&
        p.minor === min &&
        Math.abs(p.latitude - lat) < 0.00001 &&
        Math.abs(p.longitude - lng) < 0.00001
      );
    },
    [savedProfile],
  );

  // Load venue profile when bound library changes
  React.useEffect(() => {
    if (!bound?.library_id) {
      setSavedProfile(null);
      setIsSaved(false);
      return;
    }

    let isMounted = true;
    IGoService.checkin
      .getVenueProfile(bound.library_id)
      .then((profile) => {
        if (!isMounted) return;
        if (profile) {
          setSavedProfile(profile);
          setSelectedBeacon(profile.beacon_uuid);
          setMajor(profile.major);
          setMinor(profile.minor);
          setLatitude(profile.latitude);
          setLongitude(profile.longitude);
          setIsSaved(true);
          const hasInDevice = device?.beacon_uuids?.some(
            (u) => u.toUpperCase() === profile.beacon_uuid.toUpperCase(),
          );
          setCustomMode(!hasInDevice);
        } else {
          setSavedProfile(null);
          setIsSaved(false);
          setMajor(10001);
          setMinor(1980);
          setLatitude(31.2304);
          setLongitude(121.4737);
          if (device?.beacon_uuids && device.beacon_uuids.length > 0) {
            setSelectedBeacon(device.beacon_uuids[0]);
            setCustomMode(false);
          } else {
            setSelectedBeacon('');
            setCustomMode(true);
          }
        }
      })
      .catch(() => {
        if (!isMounted) return;
        setSavedProfile(null);
        setIsSaved(false);
      });

    return () => {
      isMounted = false;
    };
  }, [bound?.library_id, bound?.name, device?.beacon_uuids]);

  // Fallback if beacon list arrives after mount and beacon is empty
  React.useEffect(() => {
    if (
      !selectedBeacon &&
      !savedProfile &&
      device?.beacon_uuids &&
      device.beacon_uuids.length > 0 &&
      !customMode
    ) {
      setSelectedBeacon(device.beacon_uuids[0]);
    }
  }, [device?.beacon_uuids, selectedBeacon, savedProfile, customMode]);

  const handleBeaconChange = (val: string) => {
    setSelectedBeacon(val);
    setIsSaved(checkIfMatchesSaved(val, major, minor, latitude, longitude));
  };

  const handleMajorChange = (val: number) => {
    setMajor(val);
    setIsSaved(
      checkIfMatchesSaved(selectedBeacon, val, minor, latitude, longitude),
    );
  };

  const handleMinorChange = (val: number) => {
    setMinor(val);
    setIsSaved(
      checkIfMatchesSaved(selectedBeacon, major, val, latitude, longitude),
    );
  };

  const handleLatitudeChange = (val: number) => {
    setLatitude(val);
    setIsSaved(
      checkIfMatchesSaved(selectedBeacon, major, minor, val, longitude),
    );
  };

  const handleLongitudeChange = (val: number) => {
    setLongitude(val);
    setIsSaved(
      checkIfMatchesSaved(selectedBeacon, major, minor, latitude, val),
    );
  };

  const handleSaveProfile = async () => {
    if (!bound) {
      toast.error(t('venueNotLocked'));
      return;
    }
    const cleanBeacon = selectedBeacon.trim();
    if (!cleanBeacon) {
      toast.error('请选择或输入 Beacon UUID');
      return;
    }
    if (major < 0 || major > 65535) {
      toast.error('Major 必须介于 0 和 65535 之间');
      return;
    }
    if (minor < 0 || minor > 65535) {
      toast.error('Minor 必须介于 0 和 65535 之间');
      return;
    }
    if (latitude < -90 || latitude > 90) {
      toast.error('纬度必须介于 -90 和 90 之间');
      return;
    }
    if (longitude < -180 || longitude > 180) {
      toast.error('经度必须介于 -180 和 180 之间');
      return;
    }

    setSavingProfile(true);
    try {
      const res = await IGoService.checkin.saveVenueProfile(bound.library_id, {
        library_name: bound.name,
        beacon_uuid: cleanBeacon,
        major,
        minor,
        latitude,
        longitude,
      });
      setSavedProfile(res);
      setSelectedBeacon(res.beacon_uuid);
      setIsSaved(true);
      toast.success(t('saveSuccess'));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setSavingProfile(false);
    }
  };

  const handleSign = async () => {
    if (!bound) {
      toast.error(t('venueNotLocked'));
      return;
    }
    if (!selectedBeacon) {
      toast.error('请选择或输入 Beacon UUID');
      return;
    }

    if (
      currentReservation?.has_reservation &&
      currentReservation.library_id &&
      currentReservation.library_id !== bound.library_id
    ) {
      toast.warning(
        `当前预约在【${currentReservation.library_name}】，与当前锁定场馆【${bound.name}】不一致`,
      );
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

      // Automatically sync saved state
      const autoProfile: CheckInVenueProfile = {
        library_id: bound.library_id,
        library_name: bound.name,
        beacon_uuid: selectedBeacon.trim().toUpperCase(),
        major,
        minor,
        latitude,
        longitude,
        updated_at: new Date().toISOString(),
      };
      setSavedProfile(autoProfile);
      setIsSaved(true);
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
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <Compass className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('beaconTitle')}
            </CardTitle>
          </div>
          {bound && (
            <div className='flex items-center gap-1.5'>
              {isSaved ? (
                <Badge
                  variant='outline'
                  className='text-[10px] border-emerald-500/30 text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 font-normal shadow-none'
                >
                  <CheckCircle2 className='size-3 mr-1' />
                  {t('profileSaved')}
                </Badge>
              ) : savedProfile ? (
                <Badge
                  variant='outline'
                  className='text-[10px] border-amber-500/30 text-amber-600 dark:text-amber-400 bg-amber-500/10 font-normal shadow-none'
                >
                  {t('profileUnsaved')}
                </Badge>
              ) : (
                <Badge
                  variant='outline'
                  className='text-[10px] text-muted-foreground font-normal shadow-none'
                >
                  {t('profileEmpty')}
                </Badge>
              )}
            </div>
          )}
        </div>
        <CardDescription>
          设定打卡场馆基准 Beacon 与模拟地理坐标并执行远程签到
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        {/* 预约与场馆一致性警示 */}
        {currentReservation?.has_reservation &&
          bound &&
          currentReservation.library_id &&
          currentReservation.library_id !== bound.library_id && (
            <div className='p-2.5 rounded-lg border border-dashed border-amber-500/30 bg-amber-500/10 text-xs text-amber-900 dark:text-amber-200 flex items-center gap-2 shadow-none'>
              <AlertTriangle className='size-4 shrink-0 text-amber-600 dark:text-amber-400' />
              <span>
                当前预约在【{currentReservation.library_name}
                】，与当前目标场馆【{bound.name}】不一致，可能无法成功签到。
              </span>
            </div>
          )}

        <div className='grid grid-cols-1 sm:grid-cols-2 gap-4'>
          {/* Beacon UUID */}
          <div className='space-y-1.5'>
            <div className='flex items-center justify-between'>
              <Label className='text-xs font-medium'>{t('selectBeacon')}</Label>
              {device?.beacon_uuids && device.beacon_uuids.length > 0 && (
                <Button
                  type='button'
                  variant='ghost'
                  size='sm'
                  onClick={() => setCustomMode(!customMode)}
                  className='h-5 px-1.5 text-[11px] text-muted-foreground hover:text-foreground'
                >
                  {customMode
                    ? t('deviceBeaconToggle')
                    : t('customBeaconToggle')}
                </Button>
              )}
            </div>
            {!customMode &&
            device?.beacon_uuids &&
            device.beacon_uuids.length > 0 ? (
              <Select value={selectedBeacon} onValueChange={handleBeaconChange}>
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
                onChange={(e) => handleBeaconChange(e.target.value)}
                placeholder={t('enterCustomBeacon')}
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
              onChange={(e) => handleMajorChange(Number(e.target.value))}
              className='h-8 text-xs font-mono shadow-none bg-background'
            />
          </div>
          <div className='space-y-1'>
            <Label className='text-[11px] text-muted-foreground'>Minor</Label>
            <Input
              type='number'
              value={minor}
              onChange={(e) => handleMinorChange(Number(e.target.value))}
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
              onChange={(e) => handleLatitudeChange(Number(e.target.value))}
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
              onChange={(e) => handleLongitudeChange(Number(e.target.value))}
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

        <div className='pt-2 flex items-center gap-2.5 flex-wrap'>
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

          <Button
            type='button'
            variant='outline'
            onClick={handleSaveProfile}
            disabled={savingProfile || !bound || !selectedBeacon || isSaved}
            className='gap-1.5 shadow-none'
          >
            {savingProfile ? (
              <Spinner className='size-4' />
            ) : (
              <Save className='size-4' />
            )}
            <span>
              {savingProfile ? t('savingConfig') : t('saveConfigBtn')}
            </span>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
