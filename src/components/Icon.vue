<script lang="ts" setup>
defineProps({
  name: String,
})
</script>

<template>
  <div :class="name" />
</template>
