// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { toast } from 'sonner';
import {
  Building2,
  Check,
  ChevronDown,
  ChevronUp,
  ExternalLink,
  Radio,
  Save,
  ShieldCheck,
  Workflow,
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
  LibraryLayoutResponse,
  LibrarySummary,
  PipelineConfigDTO,
  UpdatePipelineConfigRequest,
} from '@/lib/services/igo/types';
import { cn } from '@/lib/utils';

interface PipelineEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config: PipelineConfigDTO | null;
  onSuccess: () => void;
}

const WECHAT_LOGIN_AUTH_URL =
  'https://open.weixin.qq.com/connect/oauth2/authorize?appid=wx2996d437cd442527&redirect_uri=https%3A%2F%2Fwechat.v2.traceint.com%2Findex.php%2FurlSign%2FgetUrl%3Furl%3Dhttps%253A%252F%252Fweb.traceint.com%252Fweb%252Findex.html&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect';

export function PipelineEditDialog({
  open,
  onOpenChange,
  config,
  onSuccess,
}: PipelineEditDialogProps) {
  const t = useTranslations('igo.pipeline');
  const tCommon = useTranslations('common');

  // Form states
  const [name, setName] = React.useState('');
  const [autoCheckin, setAutoCheckin] = React.useState(false);

  // Venue & Seat
  const [libraryId, setLibraryId] = React.useState<number>(0);
  const [libraryName, setLibraryName] = React.useState('');
  const [floor, setFloor] = React.useState('');
  const [seatKey, setSeatKey] = React.useState('');
  const [seatName, setSeatName] = React.useState('');

  // Beacon Settings
  const [beaconLat, setBeaconLat] = React.useState('');
  const [beaconLng, setBeaconLng] = React.useState('');
  const [beaconMac, setBeaconMac] = React.useState('');
  const [major, setMajor] = React.useState('');
  const [minor, setMinor] = React.useState('');

  // Optional Venue/Seat Picker
  const [showVenuePicker, setShowVenuePicker] = React.useState(false);
  const [venueList, setVenueList] = React.useState<LibrarySummary[]>([]);
  const [currentLayout, setCurrentLayout] =
    React.useState<LibraryLayoutResponse | null>(null);
  const [isLoadingLayout, setIsLoadingLayout] = React.useState(false);
  const [seatSearch, setSeatSearch] = React.useState('');

  // Optional Credential Update
  const [showAuthUpdate, setShowAuthUpdate] = React.useState(false);
  const [authType, setAuthType] = React.useState<'url' | 'cookie'>('url');
  const [authUrl, setAuthUrl] = React.useState('');
  const [cookie, setCookie] = React.useState('');
  const [verifiedCookie, setVerifiedCookie] = React.useState('');
  const [cookieExpiresAt, setCookieExpiresAt] = React.useState<string | null>(
    null,
  );
  const [isVerifyingAuth, setIsVerifyingAuth] = React.useState(false);

  const [isSaving, setIsSaving] = React.useState(false);

  // Initialize form from config
  React.useEffect(() => {
    if (open && config) {
      setName(config.name);
      setAutoCheckin(config.auto_checkin);
      setLibraryId(config.library_id);
      setLibraryName(config.library_name);
      setFloor(config.floor || '');
      setSeatKey(config.seat_key);
      setSeatName(config.seat_name);
      setBeaconLat(config.latitude || '');
      setBeaconLng(config.longitude || '');
      setBeaconMac(config.beacon_uuid || '');
      setMajor(
        config.major !== undefined && config.major !== null && config.major > 0
          ? String(config.major)
          : '',
      );
      setMinor(
        config.minor !== undefined && config.minor !== null && config.minor > 0
          ? String(config.minor)
          : '',
      );

      // Reset optional sections
      setShowVenuePicker(false);
      setShowAuthUpdate(false);
      setAuthUrl('');
      setCookie('');
      setVerifiedCookie('');
      setCookieExpiresAt(config.cookie_expires_at || null);
      setCurrentLayout(null);
      setSeatSearch('');
      if (config.library_id && config.library_name) {
        setVenueList([
          {
            library_id: config.library_id,
            name: config.library_name,
            floor: config.floor || '',
            is_open: true,
            total_seats: 0,
            used_seats: 0,
            booked_seats: 0,
          },
        ]);
      } else {
        setVenueList([]);
      }
    }
  }, [open, config]);

  // Optional: Verify new session credentials
  const handleVerifySession = async () => {
    const rawInput = (authType === 'cookie' ? cookie : authUrl).trim();
    if (!rawInput) {
      toast.error(
        authType === 'url' ? '请输入微信授权链接或 Code' : '请输入 Cookie',
      );
      return;
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
        }
        toast.success('TraceInt 凭据验证成功！');
      } else {
        toast.error('TraceInt 凭据验证失败，请重新获取授权链接');
      }
    } catch (err: unknown) {
      toast.error(
        err instanceof Error ? err.message : '验证 TraceInt 凭据失败',
      );
    } finally {
      setIsVerifyingAuth(false);
    }
  };

  // Fetch venue layout when picker is used
  const fetchLayoutForVenue = async (libId: number) => {
    setIsLoadingLayout(true);
    try {
      const targetCookie =
        verifiedCookie ||
        (authType === 'cookie' ? cookie : authUrl).trim() ||
        undefined;
      const layout = await IGoService.pipeline.helperGetLibraryLayout({
        cookie: targetCookie,
        library_id: libId,
      });
      setCurrentLayout(layout);
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : '获取场馆座位信息失败');
    } finally {
      setIsLoadingLayout(false);
    }
  };

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
    if (!config) return;
    if (!name.trim()) {
      toast.error('请输入配置名称');
      return;
    }
    if (!seatKey.trim()) {
      toast.error('座位编号不能为空');
      return;
    }
    if (!seatName.trim()) {
      toast.error('座位名称不能为空');
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

    const effectiveCookie =
      verifiedCookie ||
      (showAuthUpdate && (authType === 'cookie' ? cookie : authUrl).trim()) ||
      undefined;

    setIsSaving(true);
    try {
      const req: UpdatePipelineConfigRequest = {
        name: name.trim(),
        library_id: libraryId,
        library_name: libraryName,
        floor,
        seat_key: seatKey.trim(),
        seat_name: seatName.trim(),
        auto_checkin: autoCheckin,
        cookie: effectiveCookie,
        latitude: beaconLat.trim() || undefined,
        longitude: beaconLng.trim() || undefined,
        beacon_uuid: beaconMac.trim() || undefined,
        major: major.trim() !== '' ? Number(major) : 0,
        minor: minor.trim() !== '' ? Number(minor) : 0,
      };

      await IGoService.pipeline.updateConfig(config.id, req);
      toast.success('配置更新成功！');
      onSuccess();
      onOpenChange(false);
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : '保存配置失败');
    } finally {
      setIsSaving(false);
    }
  };

  if (!config) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-xl max-h-[90vh] flex flex-col p-0 gap-0 overflow-hidden'>
        <DialogHeader className='px-6 pt-6 pb-4 border-b border-border/40'>
          <div className='flex items-center gap-2'>
            <div className='flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary'>
              <Workflow className='size-4' />
            </div>
            <div>
              <DialogTitle className='text-lg font-semibold'>
                {t('dialog.editTitle')}
              </DialogTitle>
              <DialogDescription className='text-xs mt-0.5'>
                {t('dialog.editSubtitle')}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className='flex-1 overflow-y-auto px-6 py-5 space-y-5 text-sm'>
          {/* Section 1: Basic Information */}
          <div className='space-y-4'>
            <div className='grid grid-cols-2 gap-3'>
              <div className='space-y-1.5'>
                <Label
                  htmlFor='edit-config-id'
                  className='text-xs font-semibold text-muted-foreground'
                >
                  {t('dialog.idLabel')}
                </Label>
                <Input
                  id='edit-config-id'
                  value={config.id}
                  disabled
                  className='font-mono text-xs bg-muted/50 cursor-not-allowed'
                />
              </div>

              <div className='space-y-1.5'>
                <Label
                  htmlFor='edit-config-name'
                  className='text-xs font-semibold'
                >
                  {t('dialog.nameLabel')}{' '}
                  <span className='text-destructive'>*</span>
                </Label>
                <Input
                  id='edit-config-name'
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder={t('dialog.namePlaceholder')}
                  className='text-xs'
                />
              </div>
            </div>

            {/* Auto Check-in Toggle */}
            <div className='flex items-center justify-between rounded-lg border border-border/40 p-3 bg-muted/20'>
              <div className='space-y-0.5'>
                <Label
                  htmlFor='edit-auto-checkin-switch'
                  className='text-xs font-semibold cursor-pointer'
                >
                  {t('dialog.enableCheckinLabel')}
                </Label>
                <p className='text-[11px] text-muted-foreground'>
                  {t('dialog.enableCheckinDesc')}
                </p>
              </div>
              <Switch
                id='edit-auto-checkin-switch'
                checked={autoCheckin}
                onCheckedChange={setAutoCheckin}
              />
            </div>
          </div>

          {/* Section 2: Target Venue & Seat */}
          <div className='space-y-3 rounded-lg border border-border/40 p-3.5 bg-muted/10'>
            <div className='flex items-center justify-between'>
              <span className='text-xs font-semibold flex items-center gap-1.5'>
                <Building2 className='size-3.5 text-primary' />
                {t('dialog.currentVenueAndSeat')}
              </span>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                className='h-7 text-xs text-primary px-2 gap-1 hover:bg-primary/10'
                onClick={() => {
                  const next = !showVenuePicker;
                  setShowVenuePicker(next);
                  if (next && libraryId && !currentLayout) {
                    fetchLayoutForVenue(libraryId);
                  }
                }}
              >
                <span>
                  {showVenuePicker
                    ? t('dialog.hideVenueOrSeat')
                    : t('dialog.changeVenueOrSeat')}
                </span>
                {showVenuePicker ? (
                  <ChevronUp className='size-3.5' />
                ) : (
                  <ChevronDown className='size-3.5' />
                )}
              </Button>
            </div>

            <div className='grid grid-cols-2 gap-2.5'>
              <div className='space-y-1'>
                <Label className='text-[11px] text-muted-foreground'>
                  {t('card.targetVenue')}
                </Label>
                <p className='text-xs font-medium truncate p-2 rounded-md bg-background border border-border/40'>
                  {libraryName}{' '}
                  <span className='text-muted-foreground font-normal'>
                    ({floor}F)
                  </span>
                </p>
              </div>

              <div className='space-y-1'>
                <Label className='text-[11px] text-muted-foreground'>
                  {t('dialog.seatNameLabel')}
                </Label>
                <Input
                  value={seatName}
                  onChange={(e) => setSeatName(e.target.value)}
                  placeholder='如 201号'
                  className='h-8 text-xs'
                />
              </div>

              <div className='space-y-1 col-span-2'>
                <Label className='text-[11px] text-muted-foreground'>
                  {t('dialog.seatKeyLabel')}
                </Label>
                <Input
                  value={seatKey}
                  onChange={(e) => setSeatKey(e.target.value)}
                  placeholder='如 S-201'
                  className='h-8 text-xs font-mono'
                />
              </div>
            </div>

            {/* Optional Venue / Seat picker dropdown & grid */}
            {showVenuePicker && (
              <div className='space-y-3 pt-3 border-t border-border/40'>
                {venueList.length > 1 && (
                  <div className='space-y-1.5'>
                    <Label className='text-xs font-semibold'>
                      {t('dialog.selectVenueLabel')}
                    </Label>
                    <Select
                      value={String(libraryId)}
                      onValueChange={(val) => {
                        const id = Number(val);
                        setLibraryId(id);
                        const v = venueList.find(
                          (lib) => lib.library_id === id,
                        );
                        if (v) {
                          setLibraryName(v.name);
                          setFloor(v.floor);
                        }
                        fetchLayoutForVenue(id);
                      }}
                    >
                      <SelectTrigger className='h-8 text-xs'>
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
                            {lib.name} ({lib.floor}F)
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                )}

                {isLoadingLayout ? (
                  <div className='flex items-center justify-center py-6 gap-2 text-muted-foreground text-xs'>
                    <Spinner className='size-4' />
                    <span>正在拉取场馆座位图...</span>
                  </div>
                ) : currentLayout && currentLayout.seats ? (
                  <div className='space-y-2'>
                    <div className='flex items-center justify-between gap-2'>
                      <span className='text-xs font-medium'>
                        {t('dialog.selectSeatLabel')}
                      </span>
                      <Input
                        placeholder='搜索座位...'
                        value={seatSearch}
                        onChange={(e) => setSeatSearch(e.target.value)}
                        className='h-7 w-36 text-xs'
                      />
                    </div>
                    <div className='grid grid-cols-3 sm:grid-cols-4 gap-2 max-h-48 overflow-y-auto p-1 border border-border/40 rounded-lg'>
                      {availableSeats.map((seat) => {
                        const isSelected = seatKey === seat.seat_key;
                        const isOccupied = seat.is_occupied;
                        return (
                          <button
                            key={seat.seat_key}
                            type='button'
                            disabled={isOccupied}
                            onClick={() => {
                              setSeatKey(seat.seat_key);
                              setSeatName(seat.seat_name);
                            }}
                            className={cn(
                              'flex flex-col items-center justify-center p-1.5 rounded-md border text-xs transition-all relative',
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
                            {isSelected && (
                              <Check className='size-3 absolute top-1 right-1 text-primary' />
                            )}
                          </button>
                        );
                      })}
                    </div>
                  </div>
                ) : (
                  <p className='text-xs text-muted-foreground'>
                    当前无排布图缓存，可直接在上方的输入框中修改座位名称与编号。
                  </p>
                )}
              </div>
            )}
          </div>

          {/* Section 3: Beacon Settings (Required when autoCheckin is true) */}
          {autoCheckin && (
            <div className='space-y-3 rounded-lg border border-primary/20 bg-primary/5 p-3.5'>
              <div className='flex items-center justify-between'>
                <span className='text-xs font-semibold flex items-center gap-1.5'>
                  <Radio className='size-3.5 text-primary' />
                  {t('dialog.customBeaconRequired')}
                </span>
                <span className='text-[10px] text-destructive font-medium'>
                  {t('dialog.requiredTag')}
                </span>
              </div>

              <div className='grid grid-cols-2 gap-2.5'>
                <div className='space-y-1'>
                  <Label className='text-[11px] text-muted-foreground'>
                    {t('dialog.latLabel')}
                    <span className='text-destructive ml-0.5'>*</span>
                  </Label>
                  <Input
                    placeholder={t('dialog.latPlaceholder')}
                    value={beaconLat}
                    onChange={(e) => setBeaconLat(e.target.value)}
                    className='h-8 text-xs font-mono'
                  />
                </div>
                <div className='space-y-1'>
                  <Label className='text-[11px] text-muted-foreground'>
                    {t('dialog.lngLabel')}
                    <span className='text-destructive ml-0.5'>*</span>
                  </Label>
                  <Input
                    placeholder={t('dialog.lngPlaceholder')}
                    value={beaconLng}
                    onChange={(e) => setBeaconLng(e.target.value)}
                    className='h-8 text-xs font-mono'
                  />
                </div>
              </div>

              <div className='space-y-1'>
                <Label className='text-[11px] text-muted-foreground'>
                  {t('dialog.macLabel')}
                  <span className='text-destructive ml-0.5'>*</span>
                </Label>
                <Input
                  placeholder={t('dialog.macPlaceholder')}
                  value={beaconMac}
                  onChange={(e) => setBeaconMac(e.target.value)}
                  className='h-8 text-xs font-mono'
                />
              </div>

              <div className='grid grid-cols-2 gap-2.5'>
                <div className='space-y-1'>
                  <Label className='text-[11px] text-muted-foreground'>
                    {t('dialog.majorLabel')}
                    <span className='text-destructive ml-0.5'>*</span>
                  </Label>
                  <Input
                    placeholder={t('dialog.majorPlaceholder')}
                    value={major}
                    onChange={(e) => setMajor(e.target.value)}
                    className='h-8 text-xs font-mono'
                  />
                </div>
                <div className='space-y-1'>
                  <Label className='text-[11px] text-muted-foreground'>
                    {t('dialog.minorLabel')}
                    <span className='text-destructive ml-0.5'>*</span>
                  </Label>
                  <Input
                    placeholder={t('dialog.minorPlaceholder')}
                    value={minor}
                    onChange={(e) => setMinor(e.target.value)}
                    className='h-8 text-xs font-mono'
                  />
                </div>
              </div>
            </div>
          )}

          {/* Section 4: Login Credentials Status & Optional Update */}
          <div className='space-y-3 rounded-lg border border-border/40 p-3.5 bg-muted/10'>
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-2 text-xs text-emerald-700 dark:text-emerald-400 font-medium'>
                <ShieldCheck className='size-4' />
                <span>
                  {config.has_cookie || config.cookie
                    ? '已配置登录凭据 (保存时保留现有凭据)'
                    : '暂无凭据'}
                </span>
              </div>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                className='h-7 text-xs text-primary px-2 gap-1 hover:bg-primary/10'
                onClick={() => setShowAuthUpdate(!showAuthUpdate)}
              >
                <span>{t('dialog.updateAuthOptional')}</span>
                {showAuthUpdate ? (
                  <ChevronUp className='size-3.5' />
                ) : (
                  <ChevronDown className='size-3.5' />
                )}
              </Button>
            </div>

            <p className='text-[11px] text-muted-foreground'>
              {t('dialog.preserveAuthNotice')}
            </p>

            {showAuthUpdate && (
              <div className='space-y-3 pt-2 border-t border-border/40'>
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

                <div className='flex items-center gap-2'>
                  {authType === 'url' ? (
                    <Input
                      value={authUrl}
                      onChange={(e) => setAuthUrl(e.target.value)}
                      placeholder={t('dialog.authUrlPlaceholder')}
                      className='text-xs font-mono flex-1'
                    />
                  ) : (
                    <Input
                      value={cookie}
                      onChange={(e) => setCookie(e.target.value)}
                      placeholder={t('dialog.cookiePlaceholder')}
                      className='text-xs font-mono flex-1'
                    />
                  )}
                  <Button
                    type='button'
                    variant='secondary'
                    size='sm'
                    onClick={handleVerifySession}
                    disabled={
                      isVerifyingAuth || (!authUrl.trim() && !cookie.trim())
                    }
                  >
                    {isVerifyingAuth ? (
                      <Spinner className='size-3.5' />
                    ) : (
                      '验证'
                    )}
                  </Button>
                </div>

                {verifiedCookie && (
                  <div className='rounded-md bg-emerald-500/10 border border-emerald-500/20 p-2 flex items-center gap-2 text-xs text-emerald-700 dark:text-emerald-400'>
                    <ShieldCheck className='size-3.5 shrink-0' />
                    <span>
                      新凭据验证有效就绪{' '}
                      {cookieExpiresAt ? `(有效期至 ${cookieExpiresAt})` : ''}
                    </span>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        <DialogFooter className='px-6 py-4 border-t border-border/40 bg-muted/10 flex items-center justify-between sm:justify-between'>
          <Button
            type='button'
            variant='outline'
            size='sm'
            onClick={() => onOpenChange(false)}
            disabled={isSaving}
          >
            {tCommon('cancel')}
          </Button>

          <Button
            type='button'
            size='sm'
            onClick={handleSave}
            disabled={isSaving}
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
                <span>{t('dialog.saveChanges')}</span>
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
