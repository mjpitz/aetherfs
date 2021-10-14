<!--
 Copyright (C) The AetherFS Authors - All Rights Reserved
 See LICENSE for more information.
-->

<template>
  <div class="row responsive-padding double-padded" v-if="breadcrumbs.length > 0">
    <div class="responsive-padding double-padded">
      <span>
        <router-link :to="prefix">{{ root }}</router-link> /
      </span>

      <span v-for="crumb in breadcrumbs.slice(0, breadcrumbs.length - 1)" :key="crumb.key">
        <router-link :to="crumb.link">{{ crumb.label }}</router-link> /
      </span>

      <span :key="breadcrumbs[breadcrumbs.length - 1].key">
        {{ breadcrumbs[breadcrumbs.length - 1].label }}
      </span>
    </div>
  </div>

  <div class="card fluid" v-if="workingDir.children">
    <div class="section double-padded" style="background:white;" v-if="breadcrumbs.length > 0">
      <div class="row">
        <div class="col-md-12 col-lg-8">
          <router-link :to="breadcrumbs.length > 1 ? breadcrumbs[breadcrumbs.length - 1].link : prefix">..</router-link>
        </div>
      </div>
    </div>

    <div class="section double-padded" style="background:white;" v-for="file in workingDir.children" :key="file.name">
      <div class="row">
        <div class="col-md-12 col-lg-8">
          <router-link :to="`${prefix}/${file.path}`">{{ file.name }}</router-link>
        </div>
        <div style="text-align:right;" class="col-md-6 col-lg-1">{{ file.displaySize }}</div>
        <div style="text-align:right;" class="col-md-6 col-lg-3">{{ file.lastModifiedDate }}</div>
      </div>
    </div>
  </div>
</template>

<script>
import DiskSpace from "./DiskSpace.js"

export default {
  props: {
    root: {
      type: String,
    },
    prefix: {
      type: String,
    },
    path: {
      type: Array,
    },
    files: {
      type: Array,
      required: true,
    },
  },

  computed: {
    breadcrumbs() {
      const breadcrumbs = []

      let last = ''
      this.path.forEach((part) => {
        last = `${last}/${part}`

        breadcrumbs.push({
          key: last,
          label: part,
          link: this.prefix + last,
        })
      })

      return breadcrumbs
    },

    workingDir() {
      let ptr = this.fileTree

      for (let part of this.path) {
        if (!ptr || !ptr.children) {
          return {}
        }
        ptr = ptr.children[part]
      }

      return ptr || {}
    },

    fileTree() {
      const tree = {
        size: 0,
        displaySize: "",
        lastModifiedUnix: 0,
        lastModifiedDate: "",
      }

      this.files.forEach(({name, size, lastModified}) => {
        const parts = name.split('/')

        size = parseInt(size, 10)

        lastModified = new Date(lastModified)
        const lastModifiedTime = lastModified.getTime()
        const lastModifiedDate = lastModified.toDateString()

        let ptr = tree
        parts.forEach((part, i) => {
          if (!ptr.children) {
            ptr.children = {}
          }

          if (lastModifiedTime > ptr.lastModifiedUnix) {
            ptr.lastModifiedUnix = lastModifiedTime
            ptr.lastModifiedDate = lastModifiedDate
          }

          ptr.size += size
          ptr.displaySize = new DiskSpace(ptr.size).toString()

          ptr.children[part] = ptr.children[part] || {
            path: parts.slice(0, i + 1).join('/'),
            name: part,
            size: 0,
            displaySize: "",
            lastModifiedUnix: 0,
            lastModifiedDate: "",
          }

          ptr = ptr.children[part]
        })

        ptr.size = size
        ptr.displaySize = new DiskSpace(size).toString()
        ptr.lastModifiedUnix = lastModifiedTime
        ptr.lastModifiedDate = lastModifiedDate
      })

      return tree
    }
  },
}
</script>
