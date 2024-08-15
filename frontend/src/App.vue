<template>
  <div :class="['container', { collapsed: isCollapsed }]">
    <aside class="sidebar">
      <div class="menu-items">
        <h2 v-if="!isCollapsed">MediaBox</h2>
        <ul>
          <li @click="navigateTo('AlbumList')" :title="isCollapsed ? '相册' : ''">
            <span v-if="!isCollapsed">相册</span>
            <span v-else>📷</span>
          </li>
          <li @click="navigateTo('LibraryPage')" :title="isCollapsed ? '库' : ''">
            <span v-if="!isCollapsed">库</span>
            <span v-else>📚</span>
          </li>
        </ul>
      </div>
      <div class="toggle-btn" @click="toggleSidebar">
        <span v-if="!isCollapsed">⯆</span>
        <span v-else>⯈</span>
      </div>
    </aside>

    <div class="main-content">
      <router-view></router-view>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { computed } from 'vue';

const router = useRouter();
const route = useRoute();

const isLoginPage = computed(() => route.name === 'LoginForm');
const isCollapsed = ref(false);

const toggleSidebar = () => {
  isCollapsed.value = !isCollapsed.value;
};

const navigateTo = (routeName) => {
  router.push({ name: routeName });
};
</script>

<style scoped>
.container {
  display: flex;
  height: 94vh;
  /* 高度设置为视口高度，保证侧边栏铺满整个左侧 */
}

.sidebar {
  width: 200px;
  background-color: #f5f5f5;
  padding: 20px;
  box-shadow: 2px 0 5px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  transition: width 0.3s ease;
  height: 100%;
  /* 保证侧边栏的高度铺满容器 */
}

.collapsed .sidebar {
  width: 60px;
}

.menu-items {
  flex-grow: 1;
}

.sidebar .toggle-btn {
  cursor: pointer;
  text-align: center;
  padding: 10px 0;
}

.sidebar .toggle-btn span {
  font-size: 18px;
  display: inline-block;
  transform: rotate(90deg);
}

.collapsed .sidebar .toggle-btn span {
  transform: rotate(0deg);
}

.sidebar h2 {
  margin-top: 0;
  transition: opacity 0.3s ease;
}

.collapsed .sidebar h2 {
  opacity: 0;
  visibility: hidden;
}

.sidebar ul {
  list-style: none;
  padding: 0;
}

.sidebar ul li {
  cursor: pointer;
  padding: 10px 0;
  border-bottom: 1px solid #ccc;
  text-align: center;
}

.sidebar ul li:hover {
  background-color: #ddd;
}

.main-content {
  flex: 1;
  padding: 20px;
  overflow: auto;
  /* 允许内容滚动 */
}
</style>
