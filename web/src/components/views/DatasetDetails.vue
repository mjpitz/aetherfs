<!--
 Copyright (C) The AetherFS Authors - All Rights Reserved
 See LICENSE for more information.
-->

<template>
  <h1>{{ datasetFullName }}</h1>

  <div class="row">
    <div class="col-sm-3">
      <div class="responsive-padding double-padded">
        <div class="card fluid">
          <div class="section double-padded" style="background:white;">
            <router-link :to="`/dataset/${datasetFullName}`">Overview</router-link>
          </div>
        </div>
      </div>

      <div class="responsive-padding double-padded">
        <div class="card fluid">
          <div class="section double-padded" style="background:white;">
            <b>Tags</b>
          </div>
          <div class="section double-padded" style="background:white;" v-for="tag in tags" :key="tag.version">
            <router-link :to="`/dataset/${tag.name}/tag/${tag.version}`">{{ tag.version }}</router-link>
          </div>
        </div>
      </div>
    </div>

    <div class="col-sm-9">
      <div class="responsive-padding double-padded">
        <router-view :key="$route.path"></router-view>
      </div>
    </div>
  </div>
</template>

<script>
import Client from "@/api/client";

export default {
  data() {
    return {
      tags: [],
    }
  },

  mounted() {
    Client.default().ListTags(this.datasetFullName).then(({ tags }) => {
      this.tags = tags
    })
  },

  computed: {
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
