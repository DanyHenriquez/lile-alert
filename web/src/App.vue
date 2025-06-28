<script setup>
import { ref, onMounted } from 'vue'

const likeCount = ref(0)

onMounted(() => {
  const socket = new WebSocket('ws://localhost:8080/ws')
  socket.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.type === 'like_update') {
      likeCount.value = data.likes
    }
  }
})
</script>

<template>
  <div>
    <h1>YouTube Likes</h1>
    <p>Live Likes: {{ likeCount }}</p>
  </div>
</template>
