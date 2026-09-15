import { test, expect } from '@playwright/test';

test.describe('All-in-One Automation Pipeline E2E', () => {
  test.beforeEach(async ({ page, context }) => {
    // Set session and Chinese locale cookies
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

    // Mock current user
    await page.route('*/**/api/v1/user-info*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            id: '101',
            username: 'admin',
            nickname: '管理员',
            avatar_url: '',
            is_admin: true,
            email: 'admin@example.com',
          },
        }),
      });
    });

    // Mock public configs
    await page.route('*/**/api/v1/config/public*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            site_name: 'Wavelet IGo',
            allow_registration: true,
          },
        }),
      });
    });

    // Mock system settings
    await page.route('*/**/api/v1/system/settings*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: { title: 'Wavelet IGo' },
        }),
      });
    });
  });

  test('renders pipeline page and shows empty state when no cards exist', async ({
    page,
  }) => {
    await page.route('**/api/v1/igo/pipeline/configs', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: [] }),
      });
    });

    await page.goto('/pipeline');

    await expect(
      page.getByRole('heading', { name: /一条龙自动化/i }),
    ).toBeVisible();
    await expect(
      page.getByText('暂无自动化卡片', { exact: false }),
    ).toBeVisible();
    await expect(
      page.getByRole('button', { name: /新增自动化卡片/i }).first(),
    ).toBeVisible();
  });

  test('renders existing automation cards with metadata and action buttons', async ({
    page,
  }) => {
    await page.route('**/api/v1/igo/pipeline/configs', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              id: 'seat_vip_01',
              name: '考研专座-VIP',
              cookie: 'Authorization=valid-cookie-12345',
              library_id: 10,
              library_name: '总馆一楼',
              floor: '1',
              seat_key: 'S-101',
              seat_name: '101号',
              auto_checkin: true,
              checkin_token: 'token-abc',
              beacon_mac: 'AA:BB:CC:DD:EE:FF',
              created_at: '2026-09-15T08:00:00Z',
              updated_at: '2026-09-15T08:00:00Z',
            },
          ],
        }),
      });
    });

    await page.goto('/pipeline');

    await expect(page.getByText('考研专座-VIP')).toBeVisible();
    await expect(page.getByText('总馆一楼')).toBeVisible();
    await expect(page.getByText('101号')).toBeVisible();
    await expect(page.getByRole('button', { name: /立即执行/i })).toBeVisible();
  });

  test('opens 3-step creation wizard dialog and creates a new card', async ({
    page,
  }) => {
    let created = false;
    await page.route('**/api/v1/igo/pipeline/configs', async (route) => {
      if (route.request().method() === 'POST') {
        created = true;
        const postData = route.request().postDataJSON();
        expect(postData.id).toBe('my_seat_202');
        expect(postData.name).toBe('二楼靠窗202');
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            data: {
              ...postData,
              created_at: new Date().toISOString(),
            },
          }),
        });
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: created
            ? [
                {
                  id: 'my_seat_202',
                  name: '二楼靠窗202',
                  cookie: 'Authorization=cookie-xxx',
                  library_id: 20,
                  library_name: '总馆二楼',
                  floor: '2',
                  seat_key: 'S-202',
                  seat_name: '202号',
                  auto_checkin: false,
                },
              ]
            : [],
        }),
      });
    });

    await page.route('**/api/v1/igo/pipeline/verify-session', async (route) => {
      const postData = route.request().postDataJSON();
      // Enforce real backend contract: empty cookie must fail with 400 validation error
      if (
        !postData?.cookie ||
        typeof postData.cookie !== 'string' ||
        postData.cookie.trim() === ''
      ) {
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({
            error_msg: '参数校验失败',
            error: { code: 'validation_error', message: '参数校验失败' },
            data: null,
          }),
        });
        return;
      }

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            valid: true,
            libraries: [
              {
                library_id: 20,
                name: '总馆二楼',
                floor: '2',
                is_open: true,
                total_seats: 2,
                used_seats: 0,
                booked_seats: 0,
              },
              {
                library_id: 10,
                name: '总馆一楼',
                floor: '1',
                is_open: true,
                total_seats: 10,
                used_seats: 2,
                booked_seats: 1,
              },
            ],
            cookie: 'Authorization=cookie-xxx',
            expires_at: '2026-09-16T08:00:00Z',
          },
        }),
      });
    });

    await page.route('**/api/v1/igo/pipeline/library-layout', async (route) => {
      const postData = route.request().postDataJSON();
      if (
        !postData?.cookie ||
        typeof postData.cookie !== 'string' ||
        postData.cookie.trim() === ''
      ) {
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({
            error_msg: '参数校验失败',
            error: { code: 'validation_error', message: '参数校验失败' },
            data: null,
          }),
        });
        return;
      }

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            library_id: 20,
            name: '总馆二楼',
            floor: '2',
            is_open: true,
            total_seats: 2,
            available_seats: 2,
            booked_seats: 0,
            used_seats: 0,
            invalid_layout_item_count: 0,
            seats: [
              {
                seat_key: 'S-201',
                seat_name: '201号',
                is_occupied: false,
                x: 1,
                y: 1,
              },
              {
                seat_key: 'S-202',
                seat_name: '202号',
                is_occupied: false,
                x: 1,
                y: 2,
              },
            ],
          },
        }),
      });
    });

    await page.goto('/pipeline');

    // Click "新增自动化卡片"
    await page
      .getByRole('button', { name: /新增自动化卡片/i })
      .first()
      .click();
    await expect(page.getByText('新增一条龙自动化配置')).toBeVisible();

    // Step 1: Basic Info
    await page.getByPlaceholder(/例如: seat01/i).fill('my_seat_202');
    await page.getByPlaceholder(/例如: 张三的考研专座/i).fill('二楼靠窗202');

    // DO NOT switch tab: Stay on default "微信授权链接" tab!
    // Paste full real WeChat authorization redirect URL with code parameter
    const sampleWeChatAuthUrl =
      'https://web.traceint.com/web/index.html?code=081a2b3c4d5e6f7a8b9c0d1e2f3a4b5c&state=STATE';
    await page
      .getByPlaceholder(/粘贴以 open.weixin.qq.com 开头的授权链接或 Code/i)
      .fill(sampleWeChatAuthUrl);

    // Click "下一步" -> triggers verify-session with non-empty payload
    await page.getByRole('button', { name: /下一步/i }).click();

    // Step 2: Library & Seat Selection
    await expect(page.getByText('场馆与座位选择')).toBeVisible();

    // Select seat 202号
    await page.getByText('202号').click();

    // Click "下一步"
    await page.getByRole('button', { name: /下一步/i }).click();

    // Step 3: Auto Check-In
    await expect(page.getByText('自动签到设置')).toBeVisible();

    // Toggle auto-checkin switch
    await page.getByRole('switch').click();
    await page
      .getByPlaceholder(/粘贴微信签到授权链接或 Code/i)
      .fill('https://web.traceint.com/web/index.html#checkin-token-xyz');

    // Save configuration
    await page.getByRole('button', { name: /保存配置/i }).click();

    // Card should appear in list
    await expect(page.getByText('二楼靠窗202')).toBeVisible();
  });

  test('validates and rejects empty authorization input on step 1 and prevents advancing', async ({
    page,
  }) => {
    await page.route('**/api/v1/igo/pipeline/configs', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: [] }),
      });
    });

    let verifySessionCalledWithEmpty = false;
    await page.route('**/api/v1/igo/pipeline/verify-session', async (route) => {
      const postData = route.request().postDataJSON();
      if (!postData?.cookie || postData.cookie.trim() === '') {
        verifySessionCalledWithEmpty = true;
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({
            error_msg: '参数校验失败',
            error: { code: 'validation_error', message: '参数校验失败' },
            data: null,
          }),
        });
        return;
      }
      await route.fulfill({
        status: 200,
        body: JSON.stringify({ data: { valid: true } }),
      });
    });

    await page.goto('/pipeline');
    await page
      .getByRole('button', { name: /新增自动化卡片/i })
      .first()
      .click();

    // Fill ID and Name, but leave authorization input completely empty
    await page.getByPlaceholder(/例如: seat01/i).fill('empty_seat');
    await page.getByPlaceholder(/例如: 张三的考研专座/i).fill('空授权测试');

    // Click "下一步"
    await page.getByRole('button', { name: /下一步/i }).click();

    // Dialog must stay on Step 1
    await expect(page.getByText('新增一条龙自动化配置')).toBeVisible();
    await expect(page.getByText('基本信息与凭据')).toBeVisible();
    await expect(page.getByText('请输入微信授权链接或 Code')).toBeVisible();

    // Ensure frontend did not silently dispatch an empty payload to the backend
    expect(verifySessionCalledWithEmpty).toBe(false);
  });

  test('accepts raw 32-character authorization code on default URL tab without dropping payload', async ({
    page,
  }) => {
    await page.route('**/api/v1/igo/pipeline/configs', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: [] }),
      });
    });

    let receivedCode = '';
    await page.route('**/api/v1/igo/pipeline/verify-session', async (route) => {
      const postData = route.request().postDataJSON();
      receivedCode = postData?.cookie || '';
      expect(receivedCode).toBe('081a2b3c4d5e6f7a8b9c0d1e2f3a4b5c');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            valid: true,
            libraries: [
              {
                library_id: 10,
                name: '总馆一楼',
                floor: '1',
                is_open: true,
                total_seats: 10,
                used_seats: 0,
                booked_seats: 0,
              },
            ],
            cookie: 'Authorization=cookie-from-code',
          },
        }),
      });
    });

    await page.route('**/api/v1/igo/pipeline/library-layout', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            library_id: 10,
            name: '总馆一楼',
            floor: '1',
            seats: [
              {
                seat_key: 'S-1',
                seat_name: '1号',
                is_occupied: false,
                x: 1,
                y: 1,
              },
            ],
          },
        }),
      });
    });

    await page.goto('/pipeline');
    await page
      .getByRole('button', { name: /新增自动化卡片/i })
      .first()
      .click();
    await page.getByPlaceholder(/例如: seat01/i).fill('seat_code_32');
    await page.getByPlaceholder(/例如: 张三的考研专座/i).fill('32位Code测试');

    // Fill raw 32-char code in default URL input
    await page
      .getByPlaceholder(/粘贴以 open.weixin.qq.com 开头的授权链接或 Code/i)
      .fill('081a2b3c4d5e6f7a8b9c0d1e2f3a4b5c');
    await page.getByRole('button', { name: /下一步/i }).click();

    // Successfully moves to Step 2
    await expect(page.getByText('场馆与座位选择')).toBeVisible();
    expect(receivedCode).toBe('081a2b3c4d5e6f7a8b9c0d1e2f3a4b5c');
  });

  test('executes pipeline and displays step-by-step execution results dialog', async ({
    page,
  }) => {
    await page.route('**/api/v1/igo/pipeline/configs', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              id: 'seat_vip_01',
              name: '考研专座-VIP',
              cookie: 'Authorization=valid-cookie-12345',
              library_id: 10,
              library_name: '总馆一楼',
              floor: '1',
              seat_key: 'S-101',
              seat_name: '101号',
              auto_checkin: true,
            },
          ],
        }),
      });
    });

    await page.route(
      '**/api/v1/igo/pipeline/configs/seat_vip_01/run',
      async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: {
              success: true,
              session_valid: true,
              seat_available: true,
              reservation_status: '成功预约 101号 (总馆一楼)',
              checkin_status: '自动签到成功',
              message: '一条龙全流程执行成功',
            },
          }),
        });
      },
    );

    await page.goto('/pipeline');

    await page.getByRole('button', { name: /立即执行/i }).click();

    // Check Result Dialog
    await expect(
      page.getByText('自动化流程全部执行成功', { exact: false }),
    ).toBeVisible();
    await expect(page.getByText('成功预约 101号 (总馆一楼)')).toBeVisible();
    await expect(page.getByText('自动签到成功')).toBeVisible();

    // Close dialog
    await page.getByRole('button', { name: /关闭/i }).click();
  });

  test('prompts for re-authorization modal when session is expired during execution', async ({
    page,
  }) => {
    await page.route('**/api/v1/igo/pipeline/configs', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              id: 'seat_expired',
              name: '过期卡片',
              cookie: 'Authorization=expired-cookie',
              library_id: 10,
              library_name: '总馆一楼',
              seat_key: 'S-101',
              seat_name: '101号',
            },
          ],
        }),
      });
    });

    let rerunWithNewCookie = false;
    await page.route(
      '**/api/v1/igo/pipeline/configs/seat_expired/run',
      async (route) => {
        const body = route.request().postDataJSON() || {};
        if (body.cookie === 'Authorization=new-fresh-cookie') {
          rerunWithNewCookie = true;
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              data: {
                success: true,
                session_valid: true,
                seat_available: true,
                reservation_status: '成功预约 101号',
                message: '更新授权后一条龙执行成功',
              },
            }),
          });
          return;
        }
        // First run: session expired returned in payload
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: {
              success: false,
              need_auth: 'LOGIN',
              auth_url: 'https://cas.example.com/oauth/authorize',
              message: 'TraceInt 账户授权已过期，请重新登录授权',
            },
          }),
        });
      },
    );

    await page.goto('/pipeline');

    await page.getByRole('button', { name: /立即执行/i }).click();

    // Re-auth modal should pop up
    await expect(page.getByText('重新录入授权凭据')).toBeVisible();

    // Enter new cookie/code into the input
    await page
      .getByPlaceholder(/如 https:\/\/... 或 32 位 Code/i)
      .fill('Authorization=new-fresh-cookie');

    // Submit re-auth
    await page.getByRole('button', { name: /更新凭据并立即执行/i }).click();

    // Check success
    await expect(
      page.getByText('自动化流程全部执行成功', { exact: false }),
    ).toBeVisible();
    expect(rerunWithNewCookie).toBe(true);
  });

  test('handles occupied seat gracefully and displays failure in execution dialog', async ({
    page,
  }) => {
    await page.route('**/api/v1/igo/pipeline/configs', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            {
              id: 'occupied_seat',
              name: '已被占用的座位',
              cookie: 'Authorization=valid-cookie',
              library_id: 10,
              library_name: '总馆一楼',
              seat_key: 'S-101',
              seat_name: '101号',
            },
          ],
        }),
      });
    });

    await page.route(
      '**/api/v1/igo/pipeline/configs/occupied_seat/run',
      async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: {
              success: false,
              session_valid: true,
              seat_available: false,
              message: '目标座位「101号」已被他人占用，退出执行',
            },
          }),
        });
      },
    );

    await page.goto('/pipeline');

    await page.getByRole('button', { name: /立即执行/i }).click();

    // Dialog should display failure state
    await expect(page.getByText('自动化流程执行未成功')).toBeVisible();
    await expect(
      page.getByText('目标座位「101号」已被他人占用，退出执行'),
    ).toBeVisible();
  });

  test('confirms and deletes an automation card', async ({ page }) => {
    let deleted = false;
    await page.route('**/api/v1/igo/pipeline/configs', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: deleted
            ? []
            : [
                {
                  id: 'seat_to_delete',
                  name: '待删除卡片',
                  cookie: 'Authorization=valid-cookie',
                  library_id: 10,
                  library_name: '总馆一楼',
                  floor: '1',
                  seat_key: 'S-101',
                  seat_name: '101号',
                  created_at: '2026-09-15T08:00:00Z',
                  updated_at: '2026-09-15T08:00:00Z',
                },
              ],
        }),
      });
    });

    await page.route(
      '**/api/v1/igo/pipeline/configs/seat_to_delete',
      async (route) => {
        if (route.request().method() === 'DELETE') {
          deleted = true;
          await route.fulfill({
            status: 204,
          });
          return;
        }
        await route.continue();
      },
    );

    await page.goto('/pipeline');

    await expect(page.getByText('待删除卡片')).toBeVisible();

    // Open card actions dropdown
    await page.getByRole('button', { name: /操作/i }).click();
    await page.getByRole('menuitem', { name: /删除配置/i }).click();

    // Confirmation alert dialog
    await expect(page.getByText('确认删除配置卡片')).toBeVisible();
    await page.getByRole('button', { name: /删除/i }).last().click();

    // Empty state should be visible after deletion
    await expect(page.getByText('暂无自动化卡片')).toBeVisible();
  });
});
