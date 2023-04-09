<template>
  <div class="flex h-screen items-center justify-center">
    <el-form label-position="top" :model="form">
      <el-form-item label="用户名">
        <el-input v-model="form.username" />
      </el-form-item>
      <el-form-item label="邮箱">
        <el-input v-model="form.email" />
      </el-form-item>
      <el-form-item label="密码">
        <el-input v-model="form.password" />
      </el-form-item>
      <el-button type="primary" class="w-full" @click="onSignup">
        注册
      </el-button>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ElForm, ElFormItem, ElInput, ElButton } from 'element-plus'
import { ref } from 'vue'
import * as api from '@/api'

const form = ref<api.SignupWithEmailData>({
  username: '',
  email: '',
  password: '',
})

async function onSignup() {
  await api.Signup('email', form.value)
}
</script>
