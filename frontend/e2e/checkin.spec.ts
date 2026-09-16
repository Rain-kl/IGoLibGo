import { test, expect } from '@playwright/test';

test.describe('Remote check-in infos E2E', () => {
  test.beforeEach(async ({ page, context }) => {
    await context.addCookies([
      {
        name: 'wavelet_session',
        value: 'mock-auth-token-12345',
        domain: 'localhost',
        path: '/',
      },
      {
        name: 'NEXT_LOCALE',
        value: 'zh-CN',
        domain: 'localhost',
        path: '/',
      },
    ]);
    await page.route('*/**/api/v1/user-info*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            id: '101',
            username: 'admin',
            nickname: '管理员',
            is_admin: true,
          },
        }),
      });
    });
    await page.route('*/**/api/v1/config/public*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { site_name: 'Wavelet IGo' } }),
      });
    });
    await page.route('**/api/v1/igo/libraries/bound*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            bound: true,
            library: { library_id: 101, name: '主馆', floor: '2' },
          },
        }),
      });
    });
  });

  test('creates a check-in info and signs with an account', async ({
    page,
  }) => {
    const accounts = [
      {
        id: '1',
        name: '甲',
        has_cookie: true,
        has_checkin_token: true,
        nickname: '',
        school: '',
        student_name: '',
        student_number: '',
        created_at: '2026-09-16T00:00:00Z',
        updated_at: '2026-09-16T00:00:00Z',
      },
    ];
    const infos: Array<Record<string, unknown>> = [];

    await page.route('**/api/v1/igo/accounts', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: accounts }),
      });
    });
    await page.route('**/api/v1/igo/checkin/infos', async (route) => {
      if (route.request().method() === 'POST') {
        const body = route.request().postDataJSON();
        expect(body.name).toBe('主馆签到');
        const created = {
          id: '10',
          name: body.name,
          beacon_uuid: body.beacon_uuid || '',
          major: body.major || 0,
          minor: body.minor || 0,
          latitude: body.latitude || '',
          longitude: body.longitude || '',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        infos.push(created);
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({ data: created }),
        });
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: infos }),
      });
    });
    await page.route('**/api/v1/igo/checkin/infos/10/sign', async (route) => {
      const body = route.request().postDataJSON();
      expect(body.account_id).toBe('1');
      expect(body.expected_library_id).toBe(101);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { message: '验证成功' } }),
      });
    });

    await page.goto('/checkin');
    await page.getByRole('button', { name: '新建签到信息' }).click();
    await page.locator('#info-name').fill('主馆签到');
    await page.getByRole('button', { name: '保存' }).click();
    await expect(page.getByText('签到信息已保存')).toBeVisible();

    await page.getByLabel('选择账户').click();
    await page.getByRole('option', { name: '甲' }).click();
    await page.getByLabel('选择签到信息').click();
    await page.getByRole('option', { name: '主馆签到' }).click();
    await page.getByRole('button', { name: /一键远程签到打卡/ }).click();
    await expect(page.getByText('验证成功')).toBeVisible();
  });

  test('opens account check-in auth when sign returns need_auth', async ({
    page,
  }) => {
    await page.route('**/api/v1/igo/accounts', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              id: '1',
              name: '甲',
              has_cookie: true,
              has_checkin_token: false,
              nickname: '',
              school: '',
              student_name: '',
              student_number: '',
              created_at: '2026-09-16T00:00:00Z',
              updated_at: '2026-09-16T00:00:00Z',
            },
          ],
        }),
      });
    });
    await page.route('**/api/v1/igo/checkin/infos', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              id: '10',
              name: '主馆签到',
              beacon_uuid: 'FDA50693-A4E2-4FB1-AFCF-C6EB07647825',
              major: 1,
              minor: 2,
              latitude: '31.2',
              longitude: '121.4',
              created_at: '2026-09-16T00:00:00Z',
              updated_at: '2026-09-16T00:00:00Z',
            },
          ],
        }),
      });
    });
    await page.route('**/api/v1/igo/checkin/infos/**/sign', async (route) => {
      await route.fulfill({
        status: 409,
        contentType: 'application/json',
        body: JSON.stringify({
          error: {
            code: 'need_auth',
            message: '签到凭据已失效，请更新该账户的签到凭证',
          },
          data: {
            need_auth: 'CHECKIN',
            account_id: '1',
            auth_url: 'https://example.test/auth',
          },
        }),
      });
    });

    await page.goto('/checkin');
    await page.getByLabel('选择账户').click();
    await page.getByRole('option', { name: '甲' }).click();
    await page.getByLabel('选择签到信息').click();
    await page.getByRole('option', { name: '主馆签到' }).click();
    await page.locator('#sign-lib').fill('101');
    await page.getByRole('button', { name: /一键远程签到打卡/ }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(
      page.getByRole('dialog').getByRole('heading', { name: '录入签到凭据' }),
    ).toBeVisible();
    await expect(page.locator('#create-beacon-lat')).toHaveCount(0);
  });
});
