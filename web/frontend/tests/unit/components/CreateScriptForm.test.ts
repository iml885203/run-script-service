import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import CreateScriptForm from '@/components/CreateScriptForm.vue'
import { GitService } from '@/services/gitService'

// Mock GitService
vi.mock('@/services/gitService', () => ({
  GitService: {
    getGitProjects: vi.fn()
  }
}))

describe('CreateScriptForm', () => {
  const mockGitProjects = [
    {
      name: 'project1',
      path: '/home/user/project1',
      description: 'Test project 1',
      lastCommit: 'abc123'
    },
    {
      name: 'project2',
      path: '/home/user/project2',
      description: 'Test project 2'
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(GitService.getGitProjects).mockResolvedValue(mockGitProjects)
  })

  it('renders basic form elements', async () => {
    const wrapper = mount(CreateScriptForm)

    // Check basic form elements
    expect(wrapper.find('input[placeholder="Script Name"]').exists()).toBe(true)
    expect(wrapper.find('select').exists()).toBe(true) // Interval selector
    expect(wrapper.find('input[type="radio"][value="pure"]').exists()).toBe(true)
    expect(wrapper.find('input[type="radio"][value="claude-code"]').exists()).toBe(true)
  })

  it('shows pure script content textarea when pure type is selected', async () => {
    const wrapper = mount(CreateScriptForm)

    // Select pure script type
    await wrapper.find('input[type="radio"][value="pure"]').setValue(true)
    await nextTick()

    expect(wrapper.find('textarea[placeholder*="#!/bin/bash"]').exists()).toBe(true)
    expect(wrapper.find('.project-selector').exists()).toBe(false)
  })

  it('shows project selector and prompts when claude-code type is selected', async () => {
    const wrapper = mount(CreateScriptForm)

    // Select claude-code script type
    await wrapper.find('input[type="radio"][value="claude-code"]').setValue(true)
    await nextTick()

    // Should show project selector
    expect(wrapper.find('.project-selector').exists()).toBe(true)
    // Should show prompts section
    expect(wrapper.find('.prompts-container').exists()).toBe(true)
    // Should not show pure script content
    expect(wrapper.find('textarea[placeholder*="#!/bin/bash"]').exists()).toBe(false)
  })

  it('loads and displays git projects for claude-code scripts', async () => {
    const wrapper = mount(CreateScriptForm)

    // Select claude-code script type
    await wrapper.find('input[type="radio"][value="claude-code"]').setValue(true)
    await nextTick()

    // Wait for projects to load
    await nextTick()

    expect(GitService.getGitProjects).toHaveBeenCalled()

    // Check if projects are displayed
    const projectItems = wrapper.findAll('.project-item')
    expect(projectItems).toHaveLength(2)
    expect(projectItems[0].text()).toContain('project1')
    expect(projectItems[1].text()).toContain('project2')
  })

  it('allows adding and removing prompts for claude-code scripts', async () => {
    const wrapper = mount(CreateScriptForm)

    // Select claude-code script type
    await wrapper.find('input[type="radio"][value="claude-code"]').setValue(true)
    await nextTick()

    // Initially should have one prompt
    expect(wrapper.findAll('.prompt-item')).toHaveLength(1)

    // Add a prompt
    const addButton = wrapper.find('.add-prompt-btn')
    expect(addButton.exists()).toBe(true)
    await addButton.trigger('click')
    await nextTick()

    // Should now have two prompts
    expect(wrapper.findAll('.prompt-item')).toHaveLength(2)

    // Try to remove a prompt
    const removeButton = wrapper.find('.remove-prompt')
    await removeButton.trigger('click')
    await nextTick()

    // Should be back to one prompt
    expect(wrapper.findAll('.prompt-item')).toHaveLength(1)
  })

  it('limits prompts to maximum of 5', async () => {
    const wrapper = mount(CreateScriptForm)

    // Select claude-code script type
    await wrapper.find('input[type="radio"][value="claude-code"]').setValue(true)
    await nextTick()

    // Add prompts until limit
    for (let i = 0; i < 4; i++) {
      const addButton = wrapper.find('.add-prompt-btn')
      if (addButton.exists()) {
        await addButton.trigger('click')
        await nextTick()
      }
    }

    // Should have 5 prompts
    expect(wrapper.findAll('.prompt-item')).toHaveLength(5)

    // Add button should be disabled or not visible
    const addButton = wrapper.find('.add-prompt-btn')
    expect(addButton.exists()).toBe(false)
  })

  it('emits create event with correct data for pure script', async () => {
    const wrapper = mount(CreateScriptForm)

    // Fill form for pure script
    await wrapper.find('input[placeholder="Script Name"]').setValue('test-script')
    await wrapper.find('select').setValue('1h')
    await wrapper.find('input[type="radio"][value="pure"]').setValue(true)
    await nextTick()

    await wrapper.find('textarea[placeholder*="#!/bin/bash"]').setValue('echo "test"')

    // Submit form
    await wrapper.find('.btn-primary').trigger('click')
    await nextTick()

    // Check emitted event
    const createEvents = wrapper.emitted('create')
    expect(createEvents).toHaveLength(1)
    expect(createEvents?.[0][0]).toEqual({
      name: 'test-script',
      type: 'pure',
      project_path: '',
      content: 'echo "test"',
      prompts: [],
      interval: '1h',
      timeout: 0,
      max_log_lines: 100
    })
  })

  it('emits create event with correct data for claude-code script', async () => {
    const wrapper = mount(CreateScriptForm)

    // Fill form for claude-code script
    await wrapper.find('input[placeholder="Script Name"]').setValue('ai-script')
    await wrapper.find('select').setValue('30m')
    await wrapper.find('input[type="radio"][value="claude-code"]').setValue(true)
    await nextTick()

    // Wait for projects to load and select one
    await nextTick()
    const firstProject = wrapper.find('.project-item')
    await firstProject.trigger('click')

    // Add prompt
    await wrapper.find('textarea[placeholder*="Enter prompt for phase 1"]').setValue('Fix bugs')

    // Submit form
    await wrapper.find('.btn-primary').trigger('click')
    await nextTick()

    // Check emitted event
    const createEvents = wrapper.emitted('create')
    expect(createEvents).toHaveLength(1)
    expect(createEvents?.[0][0]).toEqual({
      name: 'ai-script',
      type: 'claude-code',
      project_path: '/home/user/project1',
      content: '',
      prompts: ['Fix bugs'],
      interval: '30m',
      timeout: 0,
      max_log_lines: 100
    })
  })

  it('emits cancel event when cancel button is clicked', async () => {
    const wrapper = mount(CreateScriptForm)

    await wrapper.find('.btn-secondary').trigger('click')

    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })
})
