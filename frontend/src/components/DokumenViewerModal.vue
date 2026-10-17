<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import api from '@/api/axios'

const props = defineProps<{
  isOpen: boolean
  title: string
  fileUrl: string
  fileName?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const blobUrl = ref('')
const isLoading = ref(false)
const errorMessage = ref('')

let requestId = 0
const dialog = ref<HTMLElement | null>(null)
function onKeydown(event: KeyboardEvent) { if (props.isOpen && event.key === 'Escape') emit('close') }
onMounted(() => window.addEventListener('keydown', onKeydown))

function revokeBlobUrl() {
  if (blobUrl.value) {
    URL.revokeObjectURL(blobUrl.value)
    blobUrl.value = ''
  }
}

async function loadDocument() {
  const thisRequest = ++requestId
  revokeBlobUrl()
  errorMessage.value = ''
  if (!props.isOpen || !props.fileUrl) return

  isLoading.value = true
  try {
    const response = await api.get(props.fileUrl.replace(/^\/api/, ''), { responseType: 'blob' })
    if (thisRequest !== requestId) return
    blobUrl.value = URL.createObjectURL(response.data)
  } catch {
    if (thisRequest === requestId) errorMessage.value = 'Gagal memuat preview dokumen'
  } finally {
    if (thisRequest === requestId) isLoading.value = false
  }
}

const isImage = computed(() => {
  if (!props.fileName) return false
  const ext = props.fileName.split('.').pop()?.toLowerCase()
  return ext === 'jpg' || ext === 'jpeg' || ext === 'png'
})

watch(() => [props.isOpen, props.fileUrl], async () => { await loadDocument(); if (props.isOpen) { await nextTick(); dialog.value?.focus() } }, { immediate: true })
onBeforeUnmount(() => {
  requestId++
  revokeBlobUrl()
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="emit('close')">
    <div ref="dialog" role="dialog" aria-modal="true" :aria-label="title" tabindex="-1" class="relative w-full max-w-5xl h-[85vh] bg-white rounded-xl shadow-2xl flex flex-col overflow-hidden">
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 bg-slate-50">
        <div>
          <h3 class="text-lg font-semibold text-slate-800">{{ title }}</h3>
          <p class="text-xs text-slate-500 font-mono">{{ fileName || 'Dokumen Preview' }}</p>
        </div>
        <button @click="emit('close')" aria-label="Tutup preview dokumen" class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-200 rounded-lg">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M5 6l12 12" /></svg>
        </button>
      </div>
      <div class="flex-1 w-full bg-slate-100 p-2 overflow-auto flex items-center justify-center">
        <p v-if="isLoading" class="text-sm text-slate-500">Memuat dokumen...</p>
        <p v-else-if="errorMessage" class="text-sm text-rose-600">{{ errorMessage }}</p>
        <img v-else-if="isImage && blobUrl" :src="blobUrl" :alt="title" class="max-w-full max-h-full object-contain rounded-lg shadow-sm" />
        <iframe v-else-if="blobUrl" :src="blobUrl" class="w-full h-full border-0 rounded-lg bg-white" title="Document Preview"></iframe>
      </div>
      <div class="px-6 py-3 border-t border-slate-200 bg-white flex justify-end">
        <button @click="emit('close')" class="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-700 text-sm font-medium rounded-lg">Tutup Preview</button>
      </div>
    </div>
  </div>
</template>
