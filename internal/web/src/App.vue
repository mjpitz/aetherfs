<template>
  <div class="row">
    <router-link class="button" to="/">Home</router-link>
    <router-link class="button" to="/datasets">Datasets</router-link>
  </div>

  <router-view/>
</template>

<style>
@import "~mini.css/dist/mini-default.min.css";

#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
