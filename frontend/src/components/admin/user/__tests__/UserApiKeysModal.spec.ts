import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'

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
    for (const fn of Object.values(apiMocks)) fn.mockReset()
    apiMocks.getUserApiKeys.mockResolvedValue({ items: [{ ...activeKey }] })
    apiMocks.getAllGroups.mockResolvedValue([])
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
