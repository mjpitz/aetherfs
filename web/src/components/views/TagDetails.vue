<!--
 Copyright (C) The AetherFS Authors - All Rights Reserved
 See LICENSE for more information.
-->

<template>
  <FileBrowser :root="version" :prefix="prefix" :path="path" :files="files"/>

  <pre v-if="opened">{{ opened }}</pre>

  <div class="card fluid" v-if="html">
    <div class="section double-padded" style="background:white;">
      <b>README.html</b>
    </div>
    <div class="section double-padded" style="background:white;">
      <iframe width="100%" :src="html"/>
    </div>
  </div>

  <div class="card fluid" v-if="markdown">
    <div class="section double-padded" style="background:white;">
      <b>README.md</b>
    </div>
    <div class="section double-padded" style="background:white;">
      <MarkdownView :markdown="markdown"/>
    </div>
  </div>

  <div>

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
      markdown: '',
      html: '',
    }
  },

  mounted() {
    const client = Client.default()

    client.GetDataset(this.datasetFullName, this.version).then((resp) => {
      let prefix = this.filePath
      if (prefix) {
        prefix = prefix + "/"
      }

      const readmeMarkdown = resp.dataset.files.find((file) => file.name === prefix + "README.md")
      const readmeHTML = resp.dataset.files.find((file) => file.name === prefix + "README.html")
      const openedFile = resp.dataset.files.find((file) => file.name === this.filePath)

      if (openedFile) {
        client.ReadFile(this.datasetFullName, this.version, openedFile.name).then((readme) => {
          this.opened = readme
          this.files = resp.dataset.files
        })
      } else if (readmeMarkdown) {
        client.ReadFile(this.datasetFullName, this.version, readmeMarkdown.name).then((markdown) => {
          this.markdown = markdown
          this.files = resp.dataset.files
        })
      } else if (readmeHTML) {
        this.html = client.FormatFileSystemURL(this.datasetFullName, this.version, readmeHTML.name)
        this.files = resp.dataset.files
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

<style scoped>
iframe {
  border: none;
  min-height: 600px;
  width: 134%;
  height: 100%;
  scroll-margin: 0;
  scroll-padding: 0;

  -moz-transform: scaleX(0.75) scaleY(0.9);
  -o-transform: scaleX(0.75) scaleY(0.9);
  -webkit-transform: scaleX(0.75) scaleY(0.9);
  -moz-transform-origin: 0;
  -o-transform-origin: 0;
  -webkit-transform-origin: 0;
}

@media screen and (-webkit-min-device-pixel-ratio:0) {
  iframe {
    zoom: 1;
  }
}
</style>
