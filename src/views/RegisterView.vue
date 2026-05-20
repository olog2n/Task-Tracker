<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
    <div class="w-full max-w-md bg-white rounded-xl shadow-lg p-8">
      <h1 class="text-2xl font-bold text-center text-gray-800 mb-2">Регистрация</h1>
      <p class="text-center text-gray-500 mb-6">Создайте аккаунт для начала работы</p>

      <Form v-slot="{ errors, isSubmitting }" :validation-schema="schema" @submit="onSubmit" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Имя</label>
          <Field name="name" class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" placeholder="Иван Иванов" />
          <p v-if="errors.name" class="mt-1 text-xs text-red-500">{{ errors.name }}</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <Field name="email" type="email" class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" placeholder="you@company.com" />
          <p v-if="errors.email" class="mt-1 text-xs text-red-500">{{ errors.email }}</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Пароль</label>
          <Field name="password" type="password" class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition" placeholder="••••••••" />
          <p v-if="errors.password" class="mt-1 text-xs text-red-500">{{ errors.password }}</p>
        </div>

        <div class="pt-2">
          <button type="submit" :disabled="isSubmitting" class="w-full bg-blue-600 text-white py-2.5 rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors shadow-sm">
            {{ isSubmitting ? 'Создание...' : 'Зарегистрироваться' }}
          </button>
        </div>

        <p v-if="formError" class="text-center text-sm text-red-600 bg-red-50 p-2 rounded-lg">{{ formError }}</p>
        <p class="text-center text-sm text-gray-500 mt-4">
          Уже есть аккаунт? <RouterLink to="/login" class="text-blue-600 hover:underline font-medium">Войти</RouterLink>
        </p>
      </Form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Form, Field } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import * as z from 'zod'
import { useAuth } from '@/composables/useAuth'

const { register } = useAuth()
const formError = ref('')

const schema = toTypedSchema(z.object({
  name: z.string().min(2, 'Минимум 2 символа'),
  email: z.string().email('Некорректный email'),
  password: z.string().min(6, 'Минимум 6 символов'),
}))

const onSubmit = async (values: any) => {
  formError.value = ''
  try {
    await register(values.name, values.email, values.password)
  } catch (err: any) {
    formError.value = err.response?.data?.message || 'Ошибка регистрации. Попробуйте позже.'
  }
}
</script>
