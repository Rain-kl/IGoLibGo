// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';
import {
  Check,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Radio,
  Save,
  ShieldCheck,
  Sparkles,
} from 'lucide-react';
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
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
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
  CreatePipelineConfigRequest,
  LibraryLayoutResponse,
  LibrarySummary,
} from '@/lib/services/igo/types';
import { cn } from '@/lib/utils';

interface PipelineCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

const WECHAT_LOGIN_AUTH_URL =
  'https://open.weixin.qq.com/connect/oauth2/authorize?appid=wx2996d437cd442527&redirect_uri=https%3A%2F%2Fwechat.v2.traceint.com%2Findex.php%2FurlSign%2FgetUrl%3Furl%3Dhttps%253A%252F%252Fweb.traceint.com%252Fweb%252Findex.html&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect';

export function PipelineCreateDialog({
  open,
  onOpenChange,
  onSuccess,
}: PipelineCreateDialogProps) {
  const t = useTranslations('igo.pipeline');

  const [step, setStep] = React.useState<1 | 2>(1);

  // Form states
  const [configId, setConfigId] = React.useState('');
  const [name, setName] = React.useState('');
  const [authType, setAuthType] = React.useState<'url' | 'cookie'>('url');
  const [authUrl, setAuthUrl] = React.useState('');
  const [cookie, setCookie] = React.useState('');
  const [verifiedCookie, setVerifiedCookie] = React.useState('');
  const [cookieExpiresAt, setCookieExpiresAt] = React.useState<string | null>(
    null,
  );
  const [isVerifyingAuth, setIsVerifyingAuth] = React.useState(false);

  // Auto Check-in toggle (no credentials required)
  const [autoCheckin, setAutoCheckin] = React.useState(false);

  // Beacon parameters (Required when autoCheckin is true)
  const [beaconLat, setBeaconLat] = React.useState('');
  const [beaconLng, setBeaconLng] = React.useState('');
  const [beaconMac, setBeaconMac] = React.useState('');
  const [major, setMajor] = React.useState('');
  const [minor, setMinor] = React.useState('');

  // Step 2: Venue and Seat Selection
  const [venueList, setVenueList] = React.useState<LibrarySummary[]>([]);
  const [currentLayout, setCurrentLayout] =
    React.useState<LibraryLayoutResponse | null>(null);
  const [isLoadingLayout, setIsLoadingLayout] = React.useState(false);
  const [selectedLibId, setSelectedLibId] = React.useState<number | null>(null);
  const [selectedSeatKey, setSelectedSeatKey] = React.useState('');
  const [selectedSeatName, setSelectedSeatName] = React.useState('');
  const [seatSearch, setSeatSearch] = React.useState('');

  const [isSaving, setIsSaving] = React.useState(false);

  // Reset form when dialog opens
  React.useEffect(() => {
    if (open) {
      setConfigId('');
      setName('');
      setAuthType('url');
      setAuthUrl('');
      setCookie('');
      setVerifiedCookie('');
      setCookieExpiresAt(null);
      setAutoCheckin(false);
      setBeaconLat('');
      setBeaconLng('');
      setBeaconMac('');
      setMajor('');
      setMinor('');
      setSelectedLibId(null);
      setSelectedSeatKey('');
      setSelectedSeatName('');
      setVenueList([]);
      setCurrentLayout(null);
      setSeatSearch('');
      setStep(1);
    }
  }, [open]);

  // Step 1: Verify session credentials
  const handleVerifySession = async () => {
    const rawInput = (authType === 'cookie' ? cookie : authUrl).trim();
    if (!rawInput) {
      toast.error(
        authType === 'url' ? '请输入微信授权链接或 Code' : '请输入 Cookie',
      );
      return null;
    }

    setIsVerifyingAuth(true);
    try {
      const res = await IGoService.pipeline.helperVerifySession({
        cookie: rawInput,
      });

      const isValid = Boolean(
        res &&
          (res.valid ||
            res.cookie ||
            (res.libraries && res.libraries.length > 0)),
      );

      if (isValid) {
        setVerifiedCookie(res.cookie || rawInput);
        setCookieExpiresAt(res.expires_at || null);
        if (res.libraries && res.libraries.length > 0) {
          setVenueList(res.libraries);
          if (!selectedLibId) {
            setSelectedLibId(res.libraries[0].library_id);
          }
        }
        toast.success('TraceInt 凭据验证成功！');
        return res;
      }
      toast.error('TraceInt 凭据验证失败，请重新获取授权链接');
      return null;
    } catch (err: unknown) {
      toast.error(
        err instanceof Error ? err.message : '验证 TraceInt 凭据失败',
      );
      return null;
    } finally {
      setIsVerifyingAuth(false);
    }
  };

  // Step 2: Fetch library seat layout for specific venue
  const fetchLayoutForVenue = async (libId: number, cookieToUse?: string) => {
    setIsLoadingLayout(true);
    try {
      const targetCookie =
        cookieToUse ||
        verifiedCookie ||
        (authType === 'cookie' ? cookie : authUrl).trim();
      const layout = await IGoService.pipeline.helperGetLibraryLayout({
        cookie: targetCookie || undefined,
        library_id: libId,
      });
      setCurrentLayout(layout);
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : '获取场馆座位信息失败');
    } finally {
      setIsLoadingLayout(false);
    }
  };

  const handleNextFromStep1 = async () => {
    if (!configId.trim()) {
      toast.error('请输入配置 ID');
      return;
    }
    if (!/^[a-zA-Z0-9_-]+$/.test(configId.trim())) {
      toast.error('配置 ID 只能包含英文字母、数字和下划线/连字符');
      return;
    }
    if (!name.trim()) {
      toast.error('请输入配置名称');
      return;
    }

    if (autoCheckin) {
      if (
        !beaconLat.trim() ||
        !beaconLng.trim() ||
        !beaconMac.trim() ||
        major.trim() === '' ||
        minor.trim() === ''
      ) {
        toast.error(t('dialog.beaconValidationError'));
        return;
      }
      const majorNum = Number(major);
      if (isNaN(majorNum) || majorNum < 0 || majorNum > 65535) {
        toast.error(t('dialog.majorRangeError'));
        return;
      }
      const minorNum = Number(minor);
      if (isNaN(minorNum) || minorNum < 0 || minorNum > 65535) {
        toast.error(t('dialog.minorRangeError'));
        return;
      }
    } else {
      if (major.trim() !== '') {
        const majorNum = Number(major);
        if (isNaN(majorNum) || majorNum < 0 || majorNum > 65535) {
          toast.error(t('dialog.majorRangeError'));
          return;
        }
      }
      if (minor.trim() !== '') {
        const minorNum = Number(minor);
        if (isNaN(minorNum) || minorNum < 0 || minorNum > 65535) {
          toast.error(t('dialog.minorRangeError'));
          return;
        }
      }
    }

    // Verify session
    let sessionRes = null;
    let valid = Boolean(verifiedCookie);
    const hasNewInput =
      authUrl.trim() !== '' ||
      (cookie.trim() !== '' && cookie.trim() !== verifiedCookie);

    if (!valid || hasNewInput) {
      sessionRes = await handleVerifySession();
      valid = Boolean(
        sessionRes &&
          (sessionRes.valid ||
            sessionRes.cookie ||
            (sessionRes.libraries && sessionRes.libraries.length > 0)),
      );
    }

    if (valid) {
      const availableLibs = sessionRes?.libraries || venueList;
      const targetCookie = sessionRes?.cookie || verifiedCookie;
      const targetLibId = selectedLibId || availableLibs[0]?.library_id;
      if (targetLibId) {
        if (!selectedLibId) setSelectedLibId(targetLibId);
        if (targetCookie) {
          await fetchLayoutForVenue(targetLibId, targetCookie);
        }
      }
      setStep(2);
    }
  };

  const selectedVenue = React.useMemo(() => {
    const fromList = venueList.find((lib) => lib.library_id === selectedLibId);
    if (fromList) return fromList;
    if (currentLayout && currentLayout.library_id === selectedLibId) {
      return {
        library_id: currentLayout.library_id,
        name: currentLayout.name,
        floor: currentLayout.floor,
        is_open: true,
        total_seats: 0,
        used_seats: 0,
        booked_seats: 0,
      };
    }
    return null;
  }, [venueList, currentLayout, selectedLibId]);

  const availableSeats = React.useMemo(() => {
    if (!currentLayout || !currentLayout.seats) return [];
    if (!seatSearch.trim()) return currentLayout.seats;
    const q = seatSearch.trim().toLowerCase();
    return currentLayout.seats.filter(
      (s) =>
        s.seat_name.toLowerCase().includes(q) ||
        s.seat_key.toLowerCase().includes(q),
    );
  }, [currentLayout, seatSearch]);

  const handleSave = async () => {
    if (!selectedLibId || !selectedVenue) {
      toast.error('请选择目标场馆');
      return;
    }
    if (!selectedSeatKey || !selectedSeatName) {
      toast.error('请选择目标座位');
      return;
    }

    const effectiveCookie =
      verifiedCookie ||
      (authType === 'cookie' ? cookie : authUrl).trim() ||
      undefined;

    if (!effectiveCookie) {
      toast.error('创建一条龙自动化卡片必须录入登录凭据');
      return;
    }

    setIsSaving(true);
    try {
      const req: CreatePipelineConfigRequest = {
        id: configId.trim(),
        name: name.trim(),
        library_id: selectedVenue.library_id,
        library_name: selectedVenue.name,
        floor: selectedVenue.floor,
        seat_key: selectedSeatKey,
        seat_name: selectedSeatName,
        auto_checkin: autoCheckin,
        cookie: effectiveCookie,
        latitude: beaconLat.trim() || undefined,
        longitude: beaconLng.trim() || undefined,
        beacon_uuid: beaconMac.trim() || undefined,
        major: major.trim() !== '' ? Number(major) : 0,
        minor: minor.trim() !== '' ? Number(minor) : 0,
      };
      await IGoService.pipeline.createConfig(req);
      toast.success('一条龙自动化卡片创建成功！');
      onSuccess();
      onOpenChange(false);
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : '保存配置失败');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-xl max-h-[90vh] flex flex-col p-0 gap-0 overflow-hidden'>
        <DialogHeader className='px-6 pt-6 pb-4 border-b border-border/40'>
          <div className='flex items-center gap-2'>
            <div className='flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary'>
              <Sparkles className='size-4' />
            </div>
            <div>
              <DialogTitle className='text-lg font-semibold'>
                {t('dialog.createTitle')}
              </DialogTitle>
              <DialogDescription className='text-xs mt-0.5'>
                {step === 1 && t('dialog.step1Title')}
                {step === 2 && t('dialog.step2Title')}
              </DialogDescription>
            </div>
          </div>

          {/* 2-Step Stepper Indicators */}
          <div className='flex items-center gap-2 pt-3'>
            {[1, 2].map((s) => (
              <div
                key={s}
                className={cn(
                  'h-1.5 flex-1 rounded-full transition-colors',
                  s <= step ? 'bg-primary' : 'bg-muted',
                )}
              />
            ))}
          </div>
        </DialogHeader>

        <div className='flex-1 overflow-y-auto px-6 py-5 space-y-4 text-sm'>
          {/* STEP 1: Basic Info, Credentials & Auto-Checkin */}
          {step === 1 && (
            <div className='space-y-4'>
              <div className='space-y-1.5'>
                <Label
                  htmlFor='create-config-id'
                  className='text-xs font-semibold'
                >
                  {t('dialog.idLabel')}{' '}
                  <span className='text-destructive'>*</span>
                </Label>
                <Input
                  id='create-config-id'
                  value={configId}
                  onChange={(e) => setConfigId(e.target.value)}
                  placeholder={t('dialog.idPlaceholder')}
                  className='font-mono text-sm'
                />
                <p className='text-[11px] text-muted-foreground'>
                  {t('dialog.idHelp')}
                </p>
              </div>

              <div className='space-y-1.5'>
                <Label
                  htmlFor='create-config-name'
                  className='text-xs font-semibold'
                >
                  {t('dialog.nameLabel')}{' '}
                  <span className='text-destructive'>*</span>
                </Label>
                <Input
                  id='create-config-name'
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder={t('dialog.namePlaceholder')}
                />
              </div>

              {/* Auto Check-in Switch (No credentials required) */}
              <div className='flex items-center justify-between rounded-lg border border-border/40 p-3 bg-muted/20'>
                <div className='space-y-0.5'>
                  <Label
                    htmlFor='create-auto-checkin-switch'
                    className='text-xs font-semibold cursor-pointer'
                  >
                    {t('dialog.enableCheckinLabel')}
                  </Label>
                  <p className='text-[11px] text-muted-foreground'>
                    {t('dialog.enableCheckinDesc')}
                  </p>
                </div>
                <Switch
                  id='create-auto-checkin-switch'
                  checked={autoCheckin}
                  onCheckedChange={setAutoCheckin}
                />
              </div>

              {/* Beacon Parameters (Required when autoCheckin is true) */}
              {autoCheckin && (
                <div className='space-y-3 rounded-lg border border-primary/20 bg-primary/5 p-3.5'>
                  <div className='flex items-center justify-between'>
                    <Label className='text-xs font-semibold text-foreground flex items-center gap-1.5'>
                      <Radio className='size-3.5 text-primary' />
                      <span>{t('dialog.customBeaconRequired')}</span>
                    </Label>
                    <span className='text-[10px] text-destructive font-medium'>
                      {t('dialog.requiredTag')}
                    </span>
                  </div>

                  <div className='grid grid-cols-2 gap-2.5'>
                    <div className='space-y-1'>
                      <Label
                        htmlFor='create-beacon-lat'
                        className='text-[11px] font-medium'
                      >
                        {t('dialog.latLabel')}{' '}
                        <span className='text-destructive'>*</span>
                      </Label>
                      <Input
                        id='create-beacon-lat'
                        placeholder={t('dialog.latPlaceholder')}
                        value={beaconLat}
                        onChange={(e) => setBeaconLat(e.target.value)}
                        className='h-8 text-xs font-mono'
                      />
                    </div>
                    <div className='space-y-1'>
                      <Label
                        htmlFor='create-beacon-lng'
                        className='text-[11px] font-medium'
                      >
                        {t('dialog.lngLabel')}{' '}
                        <span className='text-destructive'>*</span>
                      </Label>
                      <Input
                        id='create-beacon-lng'
                        placeholder={t('dialog.lngPlaceholder')}
                        value={beaconLng}
                        onChange={(e) => setBeaconLng(e.target.value)}
                        className='h-8 text-xs font-mono'
                      />
                    </div>
                  </div>

                  <div className='space-y-1'>
                    <Label
                      htmlFor='create-beacon-mac'
                      className='text-[11px] font-medium'
                    >
                      {t('dialog.macLabel')}{' '}
                      <span className='text-destructive'>*</span>
                    </Label>
                    <Input
                      id='create-beacon-mac'
                      placeholder={t('dialog.macPlaceholder')}
                      value={beaconMac}
                      onChange={(e) => setBeaconMac(e.target.value)}
                      className='h-8 text-xs font-mono'
                    />
                  </div>

                  <div className='grid grid-cols-2 gap-2.5'>
                    <div className='space-y-1'>
                      <Label
                        htmlFor='create-beacon-major'
                        className='text-[11px] font-medium'
                      >
                        {t('dialog.majorLabel')}{' '}
                        <span className='text-destructive'>*</span>
                      </Label>
                      <Input
                        id='create-beacon-major'
                        placeholder={t('dialog.majorPlaceholder')}
                        value={major}
                        onChange={(e) => setMajor(e.target.value)}
                        className='h-8 text-xs font-mono'
                      />
                    </div>
                    <div className='space-y-1'>
                      <Label
                        htmlFor='create-beacon-minor'
                        className='text-[11px] font-medium'
                      >
                        {t('dialog.minorLabel')}{' '}
                        <span className='text-destructive'>*</span>
                      </Label>
                      <Input
                        id='create-beacon-minor'
                        placeholder={t('dialog.minorPlaceholder')}
                        value={minor}
                        onChange={(e) => setMinor(e.target.value)}
                        className='h-8 text-xs font-mono'
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Login Credentials Section */}
              <div className='space-y-2 pt-2 border-t border-border/40'>
                <div className='flex items-center justify-between'>
                  <Label className='text-xs font-semibold'>
                    {t('dialog.authTypeLabel')}
                  </Label>
                  <a
                    href={WECHAT_LOGIN_AUTH_URL}
                    target='_blank'
                    rel='noreferrer'
                    className='text-xs text-primary hover:underline flex items-center gap-1 font-medium'
                  >
                    <span>{t('authModal.openLoginAuth')}</span>
                    <ExternalLink className='size-3' />
                  </a>
                </div>

                <Tabs
                  value={authType}
                  onValueChange={(v) => setAuthType(v as 'url' | 'cookie')}
                  className='w-full'
                >
                  <TabsList className='grid grid-cols-2 w-full'>
                    <TabsTrigger value='url' className='text-xs'>
                      {t('dialog.authTypeUrl')}
                    </TabsTrigger>
                    <TabsTrigger value='cookie' className='text-xs'>
                      {t('dialog.authTypeCookie')}
                    </TabsTrigger>
                  </TabsList>
                </Tabs>

                {authType === 'url' ? (
                  <div className='space-y-1.5'>
                    <Input
                      value={authUrl}
                      onChange={(e) => setAuthUrl(e.target.value)}
                      placeholder={t('dialog.authUrlPlaceholder')}
                      className='text-xs font-mono'
                    />
                  </div>
                ) : (
                  <div className='space-y-1.5'>
                    <Input
                      value={cookie}
                      onChange={(e) => setCookie(e.target.value)}
                      placeholder={t('dialog.cookiePlaceholder')}
                      className='text-xs font-mono'
                    />
                  </div>
                )}

                {verifiedCookie && (
                  <div className='rounded-md bg-emerald-500/10 border border-emerald-500/20 p-2.5 flex items-center gap-2 text-xs text-emerald-700 dark:text-emerald-400'>
                    <ShieldCheck className='size-4 shrink-0' />
                    <span className='flex-1 truncate'>
                      凭据有效{' '}
                      {cookieExpiresAt ? `(有效期至 ${cookieExpiresAt})` : ''}
                    </span>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* STEP 2: Venue and Seat Selection */}
          {step === 2 && (
            <div className='space-y-4'>
              {isLoadingLayout ? (
                <div className='flex flex-col items-center justify-center py-12 gap-2 text-muted-foreground'>
                  <Spinner className='size-6' />
                  <p className='text-xs'>正在拉取场馆与座位信息...</p>
                </div>
              ) : venueList.length === 0 ? (
                <div className='text-center py-8 text-muted-foreground space-y-2'>
                  <p>未获取到场馆信息，请检查登录凭据是否有效。</p>
                  <Button
                    variant='outline'
                    size='sm'
                    onClick={() => handleNextFromStep1()}
                  >
                    重试拉取
                  </Button>
                </div>
              ) : (
                <>
                  <div className='space-y-1.5'>
                    <Label className='text-xs font-semibold'>
                      {t('dialog.selectVenueLabel')}{' '}
                      <span className='text-destructive'>*</span>
                    </Label>
                    <Select
                      value={selectedLibId ? String(selectedLibId) : ''}
                      onValueChange={(val) => {
                        const id = Number(val);
                        setSelectedLibId(id);
                        setSelectedSeatKey('');
                        setSelectedSeatName('');
                        fetchLayoutForVenue(id);
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue
                          placeholder={t('dialog.selectVenuePlaceholder')}
                        />
                      </SelectTrigger>
                      <SelectContent>
                        {venueList.map((lib) => (
                          <SelectItem
                            key={lib.library_id}
                            value={String(lib.library_id)}
                          >
                            {lib.name} ({lib.floor}F) - 可选{' '}
                            {lib.total_seats -
                              (lib.used_seats + lib.booked_seats) >
                            0
                              ? lib.total_seats -
                                (lib.used_seats + lib.booked_seats)
                              : 0}
                            /{lib.total_seats || 0}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {selectedVenue && (
                    <div className='space-y-2 pt-2 border-t border-border/40'>
                      <div className='flex items-center justify-between gap-2'>
                        <Label className='text-xs font-semibold'>
                          {t('dialog.selectSeatLabel')}{' '}
                          <span className='text-destructive'>*</span>
                        </Label>
                        <Input
                          placeholder='搜索座位编号/名称...'
                          value={seatSearch}
                          onChange={(e) => setSeatSearch(e.target.value)}
                          className='h-7 w-44 text-xs'
                        />
                      </div>

                      <div className='grid grid-cols-3 sm:grid-cols-4 gap-2 max-h-56 overflow-y-auto p-1 border border-border/40 rounded-lg'>
                        {availableSeats.map((seat) => {
                          const isSelected = selectedSeatKey === seat.seat_key;
                          const isOccupied = seat.is_occupied;
                          return (
                            <button
                              key={seat.seat_key}
                              type='button'
                              disabled={isOccupied}
                              onClick={() => {
                                setSelectedSeatKey(seat.seat_key);
                                setSelectedSeatName(seat.seat_name);
                              }}
                              className={cn(
                                'flex flex-col items-center justify-center p-2 rounded-md border text-xs transition-all relative',
                                isSelected
                                  ? 'border-primary bg-primary/10 font-semibold text-primary'
                                  : isOccupied
                                    ? 'border-border/30 bg-muted/30 text-muted-foreground/50 cursor-not-allowed'
                                    : 'border-border/60 hover:border-primary/40 hover:bg-accent cursor-pointer',
                              )}
                            >
                              <span className='truncate max-w-full'>
                                {seat.seat_name}
                              </span>
                              <span className='text-[10px] text-muted-foreground font-mono'>
                                {seat.seat_key}
                              </span>
                              {isOccupied && (
                                <span className='text-[9px] text-rose-500 font-normal'>
                                  已占
                                </span>
                              )}
                              {isSelected && (
                                <Check className='size-3 absolute top-1 right-1 text-primary' />
                              )}
                            </button>
                          );
                        })}
                      </div>

                      {selectedSeatKey && (
                        <p className='text-xs text-muted-foreground'>
                          当前选定：
                          <span className='font-semibold text-foreground'>
                            {selectedSeatName}
                          </span>{' '}
                          ({selectedSeatKey})
                        </p>
                      )}
                    </div>
                  )}
                </>
              )}
            </div>
          )}
        </div>

        <DialogFooter className='px-6 py-4 border-t border-border/40 bg-muted/10 flex items-center justify-between sm:justify-between'>
          {step === 2 ? (
            <Button
              type='button'
              variant='outline'
              size='sm'
              onClick={() => setStep(1)}
              disabled={isSaving}
            >
              <ChevronLeft className='size-4 mr-1' />
              {t('dialog.prevStep')}
            </Button>
          ) : (
            <div />
          )}

          {step === 1 ? (
            <Button
              type='button'
              size='sm'
              onClick={handleNextFromStep1}
              disabled={isVerifyingAuth}
            >
              {isVerifyingAuth ? (
                <>
                  <Spinner className='size-3.5 mr-1' />
                  <span>{t('dialog.verifying')}</span>
                </>
              ) : (
                <>
                  <span>{t('dialog.nextStep')}</span>
                  <ChevronRight className='size-4 ml-1' />
                </>
              )}
            </Button>
          ) : (
            <Button
              type='button'
              size='sm'
              onClick={handleSave}
              disabled={isSaving || !selectedLibId || !selectedSeatKey}
              className='gap-1.5'
            >
              {isSaving ? (
                <>
                  <Spinner className='size-3.5' />
                  <span>{t('dialog.saving')}</span>
                </>
              ) : (
                <>
                  <Save className='size-4' />
                  <span>{t('dialog.save')}</span>
                </>
              )}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
