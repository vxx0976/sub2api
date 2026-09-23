import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import type { AdminUser } from '@/types'

const apiMocks = vi.hoisted(() => ({
  getUserApiKeys: vi.fn(),
  getAllGroups: vi.fn(),
  updateApiKeyStatus: vi.fn(),
  updateApiKeyGroup: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: { getUserApiKeys: apiMocks.getUserApiKeys },
    groups: { getAll: apiMocks.getAllGroups },
    apiKeys: {
      updateApiKeyStatus: apiMocks.updateApiKeyStatus,
      updateApiKeyGroup: apiMocks.updateApiKeyGroup
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showSuccess: apiMocks.showSuccess, showError: apiMocks.showError })
}))

vi.mock('@/utils/format', () => ({ formatDateTime: (value: string) => value }))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

vi.mock('@/components/common/BaseDialog.vue', () => ({
  default: {
    name: 'BaseDialog',
    props: ['show', 'title', 'width'],
    template: '<div v-if="show"><slot /><slot name="footer" /></div>'
  }
}))

import UserApiKeysModal from '../UserApiKeysModal.vue'

enableAutoUnmount(afterEach)

beforeEach(() => {
  for (const fn of Object.values(apiMocks)) fn.mockReset()
  apiMocks.getAllGroups.mockResolvedValue([])
  vi.spyOn(console, 'error').mockImplementation(() => {})
})

afterEach(() => vi.restoreAllMocks())

const activeKey = {
  id: 10,
  name: 'cli',
  key: 'sk-abcdefghijklmnopqrstuvwxyz012345',
  status: 'active',
  group_id: null,
  group: null,
  created_at: '2026-09-01T00:00:00Z'
}

async function mountAndOpen() {
  const wrapper = mount(UserApiKeysModal, {
    props: { show: false, user: { id: 1, email: 'u@example.com', username: 'u' } as any },
    global: { stubs: { GroupBadge: true, GroupOptionItem: true, Teleport: true } }
  })
  await wrapper.setProps({ show: true })
  await flushPromises()
  return wrapper
}

describe('UserApiKeysModal 单把 key 启停', () => {
  beforeEach(() => {
    apiMocks.getUserApiKeys.mockResolvedValue({ items: [{ ...activeKey }] })
  })

  it('停用按钮把该 key 置为 inactive，并只影响这一把 key', async () => {
    apiMocks.updateApiKeyStatus.mockResolvedValue({ api_key: { ...activeKey, status: 'inactive' } })
    const wrapper = await mountAndOpen()

    await wrapper.get('[data-test="toggle-key-status-10"]').trigger('click')
    await flushPromises()

    expect(apiMocks.updateApiKeyStatus).toHaveBeenCalledWith(10, 'inactive')
    expect(wrapper.get('[data-test="toggle-key-status-10"]').text()).toBe('admin.users.enableKey')
    expect(apiMocks.showSuccess).toHaveBeenCalledWith('admin.users.keyDisabledSuccess')
    wrapper.unmount()
  })

  it('已停用的 key 再点是启用', async () => {
    apiMocks.getUserApiKeys.mockResolvedValue({ items: [{ ...activeKey, status: 'inactive' }] })
    apiMocks.updateApiKeyStatus.mockResolvedValue({ api_key: { ...activeKey, status: 'active' } })
    const wrapper = await mountAndOpen()

    await wrapper.get('[data-test="toggle-key-status-10"]').trigger('click')
    await flushPromises()

    expect(apiMocks.updateApiKeyStatus).toHaveBeenCalledWith(10, 'active')
    wrapper.unmount()
  })

  it('额度耗尽等自动状态不给手工开关', async () => {
    apiMocks.getUserApiKeys.mockResolvedValue({ items: [{ ...activeKey, status: 'quota_exhausted' }] })
    const wrapper = await mountAndOpen()

    expect(wrapper.find('[data-test="toggle-key-status-10"]').exists()).toBe(false)
    wrapper.unmount()
  })
})

function deferred() {
  let resolve!: (value: unknown) => void
  let reject!: (error: Error) => void
  const promise = new Promise((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}
const user = (id: number) => ({ id, email: `user${id}@example.com`, username: `user${id}` }) as AdminUser
const keys = (id: number, name: string) => ({ items: [{ id, name, key: 'sk-example-key-value-for-tests', status: 'active', created_at: '2026-09-20', group_id: null }] })
async function open() {
  const wrapper = mount(UserApiKeysModal, {
    props: { show: false, user: user(1) },
    global: { stubs: { GroupBadge: true, GroupOptionItem: true } },
  })
  await wrapper.setProps({ show: true })
  return wrapper
}
async function switchUser(wrapper: Awaited<ReturnType<typeof open>>) {
  await wrapper.setProps({ show: false })
  await wrapper.setProps({ show: true, user: user(2) })
}

describe('user API key loading', () => {
  const getKeys = apiMocks.getUserApiKeys

  it('does not display the previous user keys when the next load fails', async () => {
    getKeys.mockResolvedValueOnce(keys(1, 'first-user-key')).mockRejectedValueOnce(new Error('unavailable'))
    const wrapper = await open(); await flushPromises()
    expect(wrapper.text()).toContain('first-user-key')
    await switchUser(wrapper); await flushPromises()
    expect(wrapper.text()).toContain('user2@example.com')
    expect(wrapper.text()).not.toContain('first-user-key')
  })

  it('does not replace current keys with a late previous response', async () => {
    const old = deferred()
    getKeys.mockReturnValueOnce(old.promise).mockResolvedValueOnce(keys(2, 'current-user-key'))
    const wrapper = await open()
    await switchUser(wrapper); await flushPromises()
    old.resolve(keys(1, 'old-user-key')); await flushPromises()
    expect(wrapper.text()).toContain('current-user-key')
    expect(wrapper.text()).not.toContain('old-user-key')
  })

  it('keeps the current request loading when an obsolete request fails', async () => {
    const old = deferred(); const current = deferred()
    getKeys.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const wrapper = await open()
    await switchUser(wrapper)
    old.reject(new Error('obsolete')); await flushPromises()
    expect(wrapper.find('.animate-spin').exists()).toBe(true)
    current.resolve(keys(2, 'current-user-key')); await flushPromises()
    expect(wrapper.text()).toContain('current-user-key')
    expect(wrapper.find('.animate-spin').exists()).toBe(false)
  })

  it('loads keys when the selected user changes while the dialog is open', async () => {
    getKeys.mockResolvedValueOnce(keys(1, 'first-user-key')).mockResolvedValueOnce(keys(2, 'second-user-key'))
    const wrapper = await open(); await flushPromises()
    await wrapper.setProps({ user: user(2) }); await flushPromises()
    expect(getKeys).toHaveBeenLastCalledWith(2)
    expect(wrapper.text()).toContain('second-user-key')
    expect(wrapper.text()).not.toContain('first-user-key')
  })
})
