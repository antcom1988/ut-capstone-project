<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  isOpen: boolean
  title: string
  fileUrl: string
  fileName?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const isImage = computed(() => {
  if (!props.fileName) return false
  const ext = props.fileName.split('.').pop()?.toLowerCase()
  return ext === 'jpg' || ext === 'jpeg' || ext === 'png'
})
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
    <div class="relative w-full max-w-5xl h-[85vh] bg-white rounded-xl shadow-2xl flex flex-col overflow-hidden">
      <!-- Modal Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 bg-slate-50">
        <div>
          <h3 class="text-lg font-semibold text-slate-800">{{ title }}</h3>
          <p class="text-xs text-slate-500 font-mono">{{ fileName || 'Dokumen Preview' }}</p>
        </div>
        <button
          @click="emit('close')"
          class="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-200 rounded-lg transition-colors"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Modal Body (Preview Viewer Tanpa Download) -->
      <div class="flex-1 w-full bg-slate-100 p-2 overflow-auto flex items-center justify-center">
        <!-- Jika Gambar -->
        <img
          v-if="isImage"
          :src="fileUrl"
          :alt="title"
          class="max-w-full max-h-full object-contain rounded-lg shadow-sm"
        />
        <!-- Jika PDF / Dokumen Dokumen Inline -->
        <iframe
          v-else
          :src="fileUrl"
          class="w-full h-full border-0 rounded-lg bg-white"
          title="Document Preview"
        ></iframe>
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-3 border-t border-slate-200 bg-white flex justify-end">
        <button
          @click="emit('close')"
          class="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-700 text-sm font-medium rounded-lg transition-colors"
        >
          Tutup Preview
        </button>
      </div>
    </div>
  </div>
</template>
