<!--
 Copyright (C) The AetherFS Authors - All Rights Reserved
 See LICENSE for more information.
-->

<template>
  <h1>Datasets</h1>

  <div class="row">
    <div class="card" v-for="dataset in datasets" :key="dataset.name">
      <div class="section double-padded" style="background:white;">
        <div class="responsive-padding" >
          <router-link :to="`/dataset/${dataset.name}`">{{ dataset.name }}</router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Client from "../../api/client.js"

export default {
  data() {
    return {
      datasets: [],
    }
  },

  mounted() {
    Client.default().ListDatasets().then((resp) => {
      this.datasets = resp.datasets
    })
  },
}
</script>
