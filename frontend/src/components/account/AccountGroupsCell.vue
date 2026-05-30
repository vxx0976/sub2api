<template>
  <div v-if="groups && groups.length > 0" class="flex flex-col items-start gap-1 max-w-56">
    <!-- 分组容器：每个标签独占一行,不限制数量 -->
    <GroupBadge
      v-for="group in sortedGroups"
      :key="group.id"
      :name="group.name"
      :platform="group.platform"
      :subscription-type="group.subscription_type"
      :rate-multiplier="group.rate_multiplier"
      :show-rate="false"
      class="max-w-full"
    />
  </div>
  <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import type { Group } from '@/types'

interface Props {
  groups: Group[] | null | undefined
}

const props = defineProps<Props>()

// 按 sort_order 降序排序的分组
const sortedGroups = computed(() => {
  if (!props.groups) return []
  return [...props.groups].sort((a, b) => (b.sort_order || 0) - (a.sort_order || 0))
})
</script>
