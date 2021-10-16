<!--
 Copyright (C) The AetherFS Authors - All Rights Reserved
 See LICENSE for more information.
-->

<template>
  <div class="container">
    <div class="row">
      <div class="col-sm-5 col-sm-offset-1">
        <router-link id="brand" class="secondary button" to="/">AetherFS</router-link>
        <a class="button" :href="`${baseUrl}swagger/`">Swagger</a>
      </div>
      <div class="col-sm-5" style="text-align: right;">
<!--        <router-link class="button" to="/">Log In</router-link>-->
      </div>
    </div>

    <div class="row">
      <div class="col-sm-10 col-sm-offset-1">
        <router-view/>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      baseUrl: "",
    }
  },

  mounted() {
    this.baseUrl = process.env.BASE_URL
  }
}
</script>

<style scoped>
@import "~normalize.css/normalize.css";
@import "~mini.css/dist/mini-default.min.css";

#brand {
  font-weight: bold;
}
</style>
