import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import Scripts from '@/views/Scripts.vue'
import { ref } from 'vue'

// Mock useScripts composable
const mockUseScripts = vi.fn()

vi.mock('@/composables/useScripts', () => ({
  useScripts: () => mockUseScripts()
}))

describe('Scripts Component', () => {
  let router: any

  beforeEach(() => {
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', component: { template: '<div>Home</div>' } },
        { path: '/scripts', component: Scripts }
      ]
    })

    // Default mock implementation
    mockUseScripts.mockReturnValue({
      scripts: ref([
        {
          name: 'test1',
          path: '/home/logan/run-script-service-develop/test1.sh',
          interval: 300,
          enabled: true,
          timeout: 0
        },
        {
          name: 'test2',
          path: '/home/logan/run-script-service-develop/test2.sh',
          interval: 600,
          enabled: true,
          timeout: 30
        }
      ]),
      loading: ref(false),
      error: ref(null),
      fetchScripts: vi.fn(),
      addScript: vi.fn(),
      runScript: vi.fn(),
      toggleScript: vi.fn(),
      deleteScript: vi.fn(),
      updateScript: vi.fn(),
      updateScript: vi.fn()
    })
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should call fetchScripts on mount', () => {
    const mockFetchScripts = vi.fn()
    mockUseScripts.mockReturnValue({
      scripts: ref([]),
      loading: ref(false),
      error: ref(null),
      fetchScripts: mockFetchScripts,
      addScript: vi.fn(),
      runScript: vi.fn(),
      toggleScript: vi.fn(),
      deleteScript: vi.fn(),
      updateScript: vi.fn(),
      updateScript: vi.fn()
    })

    mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    // This should pass - the component should call fetchScripts on mount
    expect(mockFetchScripts).toHaveBeenCalledOnce()
  })

  it('should display script cards with complete information', () => {
    const wrapper = mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    // Check that script cards are rendered with data-testid
    const scriptCards = wrapper.findAll('[data-testid="script-card"]')
    expect(scriptCards).toHaveLength(2)

    // Check first script details
    const firstCard = scriptCards[0]
    expect(firstCard.find('[data-testid="script-name"]').text()).toBe('test1')
    expect(firstCard.find('[data-testid="script-path"]').text()).toBe('/home/logan/run-script-service-develop/test1.sh')
    expect(firstCard.find('[data-testid="script-interval"]').text()).toBe('Interval: 300s')
    expect(firstCard.find('[data-testid="script-status"]').text()).toContain('Enabled')

    // Check second script details
    const secondCard = scriptCards[1]
    expect(secondCard.find('[data-testid="script-name"]').text()).toBe('test2')
    expect(secondCard.find('[data-testid="script-path"]').text()).toBe('/home/logan/run-script-service-develop/test2.sh')
    expect(secondCard.find('[data-testid="script-interval"]').text()).toBe('Interval: 600s')
    expect(secondCard.find('[data-testid="script-status"]').text()).toContain('Enabled')
  })

  it('should display no scripts message when scripts array is empty', () => {
    mockUseScripts.mockReturnValue({
      scripts: ref([]),
      loading: ref(false),
      error: ref(null),
      fetchScripts: vi.fn(),
      addScript: vi.fn(),
      runScript: vi.fn(),
      toggleScript: vi.fn(),
      deleteScript: vi.fn(),
      updateScript: vi.fn()
    })

    const wrapper = mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    const noScriptsMessage = wrapper.find('[data-testid="no-scripts-message"]')
    expect(noScriptsMessage.exists()).toBe(true)
    expect(noScriptsMessage.text()).toContain('No scripts configured yet')
  })

  it('should show loading state', () => {
    mockUseScripts.mockReturnValue({
      scripts: ref([]),
      loading: ref(true),
      error: ref(null),
      fetchScripts: vi.fn(),
      addScript: vi.fn(),
      runScript: vi.fn(),
      toggleScript: vi.fn(),
      deleteScript: vi.fn(),
      updateScript: vi.fn()
    })

    const wrapper = mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.find('.loading').exists()).toBe(true)
    expect(wrapper.text()).toContain('Loading scripts...')
  })

  it('should show error state', () => {
    mockUseScripts.mockReturnValue({
      scripts: ref([]),
      loading: ref(false),
      error: ref('Failed to fetch scripts'),
      fetchScripts: vi.fn(),
      addScript: vi.fn(),
      runScript: vi.fn(),
      toggleScript: vi.fn(),
      deleteScript: vi.fn(),
      updateScript: vi.fn()
    })

    const wrapper = mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    expect(wrapper.find('.error').exists()).toBe(true)
    expect(wrapper.text()).toContain('Failed to fetch scripts')
  })

  it('should show edit modal when edit button is clicked', async () => {
    const wrapper = mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    // Find and click the edit button for the first script
    const editButton = wrapper.findAll('.btn-secondary').find(btn => btn.text() === 'Edit')
    expect(editButton).toBeDefined()

    await editButton!.trigger('click')

    // Check that edit modal is shown
    const editModal = wrapper.find('[data-testid="edit-modal"]')
    expect(editModal.exists()).toBe(true)

    // Check that form is pre-populated with script data
    const nameInput = wrapper.find('[data-testid="edit-name-input"]')
    const pathInput = wrapper.find('[data-testid="edit-path-input"]')
    const intervalInput = wrapper.find('[data-testid="edit-interval-input"]')

    expect(nameInput.element.value).toBe('test1')
    expect(pathInput.element.value).toBe('/home/logan/run-script-service-develop/test1.sh')
    expect(intervalInput.element.value).toBe('300')
  })

  it('should call updateScript when edit form is submitted', async () => {
    const mockUpdateScript = vi.fn()
    mockUseScripts.mockReturnValue({
      scripts: ref([
        {
          name: 'test1',
          path: '/home/logan/run-script-service-develop/test1.sh',
          interval: 300,
          enabled: true,
          timeout: 0
        }
      ]),
      loading: ref(false),
      error: ref(null),
      fetchScripts: vi.fn(),
      addScript: vi.fn(),
      runScript: vi.fn(),
      toggleScript: vi.fn(),
      deleteScript: vi.fn(),
      updateScript: mockUpdateScript
    })

    const wrapper = mount(Scripts, {
      global: {
        plugins: [router]
      }
    })

    // Click edit button
    const editButton = wrapper.findAll('.btn-secondary').find(btn => btn.text() === 'Edit')
    await editButton!.trigger('click')

    // Modify form data
    const pathInput = wrapper.find('[data-testid="edit-path-input"]')
    await pathInput.setValue('/new/path/script.sh')

    const intervalInput = wrapper.find('[data-testid="edit-interval-input"]')
    await intervalInput.setValue('600')

    // Submit form
    const editForm = wrapper.find('[data-testid="edit-form"]')
    await editForm.trigger('submit.prevent')

    // Verify updateScript was called with correct data
    expect(mockUpdateScript).toHaveBeenCalledWith('test1', {
      path: '/new/path/script.sh',
      interval: 600,
      enabled: true,
      timeout: 0
    })
  })
})
