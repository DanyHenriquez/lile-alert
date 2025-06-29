
<template>
  <div class="h-screen flex items-center justify-center">
    <!-- Bottom-left corner bar -->
    <div
      class="fixed bottom-0 left-0 flex items-center
             border border-l-0 border-t-0 border-purple-800
             rounded-r-lg rounded-t-lg
             shadow-lg
             max-w-[320px] min-w-[220px] h-14
             overflow-hidden
             px-4">

      <div class="pl-4 flex-1">
        <p class="text-lavender-light font-medium truncate">
          Likes: 1,234&nbsp;&nbsp;&bull;&nbsp;&nbsp;
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Define a custom lighter lavender for font */
.text-lavender-light {
  color: #c8a2c8; /* soft lavender */
}
</style>


