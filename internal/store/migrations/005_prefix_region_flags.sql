UPDATE upstream_nodes
SET region = CASE region
  WHEN '美国' THEN '🇺🇸美国'
  WHEN '加拿大' THEN '🇨🇦加拿大'
  WHEN '土耳其' THEN '🇹🇷土耳其'
  WHEN '俄罗斯' THEN '🇷🇺俄罗斯'
  WHEN '越南' THEN '🇻🇳越南'
  WHEN '印尼' THEN '🇮🇩印尼'
  WHEN '日本' THEN '🇯🇵日本'
  WHEN '韩国' THEN '🇰🇷韩国'
  WHEN '新加坡' THEN '🇸🇬新加坡'
  WHEN '澳洲' THEN '🇦🇺澳洲'
  WHEN '阿联酋' THEN '🇦🇪阿联酋'
  WHEN '印度' THEN '🇮🇳印度'
  WHEN '德国' THEN '🇩🇪德国'
  WHEN '英国' THEN '🇬🇧英国'
  WHEN '巴西' THEN '🇧🇷巴西'
  WHEN '智利' THEN '🇨🇱智利'
  WHEN '法国' THEN '🇫🇷法国'
  WHEN '墨西哥' THEN '🇲🇽墨西哥'
  WHEN '荷兰' THEN '🇳🇱荷兰'
  WHEN '以色列' THEN '🇮🇱以色列'
  WHEN '西班牙' THEN '🇪🇸西班牙'
  WHEN '阿根廷' THEN '🇦🇷阿根廷'
  WHEN '乌克兰' THEN '🇺🇦乌克兰'
  WHEN '瑞士' THEN '🇨🇭瑞士'
  WHEN '南非' THEN '🇿🇦南非'
  WHEN '马来西亚' THEN '🇲🇾马来西亚'
  WHEN '菲律宾' THEN '🇵🇭菲律宾'
  WHEN '泰国' THEN '🇹🇭泰国'
  ELSE region
END,
updated_at = CURRENT_TIMESTAMP
WHERE region IN (
  '美国',
  '加拿大',
  '土耳其',
  '俄罗斯',
  '越南',
  '印尼',
  '日本',
  '韩国',
  '新加坡',
  '澳洲',
  '阿联酋',
  '印度',
  '德国',
  '英国',
  '巴西',
  '智利',
  '法国',
  '墨西哥',
  '荷兰',
  '以色列',
  '西班牙',
  '阿根廷',
  '乌克兰',
  '瑞士',
  '南非',
  '马来西亚',
  '菲律宾',
  '泰国'
);
