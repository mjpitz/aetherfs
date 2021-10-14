<!--
 Copyright (C) The AetherFS Authors - All Rights Reserved
 See LICENSE for more information.
-->

<template>
  <div v-html="compiledMarkdown"/>
</template>

<script>
import marked from "marked"

export default {
  props: {
    markdown: {
      type: String,
      required: true,
    },
  },

  computed: {
    compiledMarkdown() {
      return marked(this.markdown)
    },
  },
}
</script>
