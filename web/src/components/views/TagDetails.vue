<!--
 Copyright (C) The AetherFS Authors - All Rights Reserved
 See LICENSE for more information.
-->

<template>
  <FileBrowser :root="version" :prefix="prefix" :path="path" :files="files"/>

  <pre v-if="opened">{{ opened }}</pre>

  <div class="card fluid" v-if="readme">
    <div class="section double-padded" style="background:white;">
      <b>README.md</b>
    </div>
    <div class="section double-padded" style="background:white;">
      <MarkdownView :markdown="readme"/>
    </div>
  </div>
</template>

<script>
import Client from "@/api/client.js"
import FileBrowser from "@/components/files/FileBrowser"
import MarkdownView from "@/components/markdown/MarkdownView.vue"

export default {
  components: {
    FileBrowser,
    MarkdownView,
  },

  data() {
    return {
      files: [],
      opened: '',
      readme: '',
    }
  },

  mounted() {
    const client = Client.default()

    client.GetDataset(this.datasetFullName, this.version).then((resp) => {
      let prefix = this.filePath
      if (prefix) {
        prefix = prefix + "/"
      }

      const readmeFileName = prefix + "README.md"
      const readmeFile = resp.dataset.files.find((file) => file.name === readmeFileName)
      const openedFile = resp.dataset.files.find((file) => file.name === this.filePath)

      if (openedFile) {
        client.ReadFile(this.datasetFullName, this.version, openedFile.name).then((readme) => {
          this.opened = readme
          this.files = resp.dataset.files
        })
      } else if (readmeFile) {
        client.ReadFile(this.datasetFullName, this.version, readmeFile.name).then((readme) => {
          this.readme = readme
          this.files = resp.dataset.files
        })
      } else {
        this.files = resp.dataset.files
      }
    })
  },

  computed: {
    prefix() {
      return `/dataset/${this.datasetFullName}/tag/${this.version}/tree`
    },

    path() {
      return this.$route.params.path || []
    },

    filePath() {
      return this.path.join('/')
    },

    version() {
      return this.$route.params.version
    },

    datasetFullName() {
      let {scope, dataset} = this.$route.params
      if (scope) {
        dataset = scope + "/" + dataset
      }
      return dataset
    }
  }
}
</script>