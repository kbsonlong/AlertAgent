/**
 * 格式化日期时间
 * @param dateStr ISO格式的日期字符串
 * @returns 格式化后的日期时间字符串
 */
export const formatDateTime = (dateStr: string): string => {
  if (!dateStr) return '-';
  
  const date = new Date(dateStr);
  if (isNaN(date.getTime())) {
    return '-';
  }

  try {
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false
    });
  } catch {
    return '-';
  }
}; 