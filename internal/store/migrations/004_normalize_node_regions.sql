UPDATE upstream_nodes
SET region = '美国',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('美国', '美西', '美东', 'US', 'us', 'USA', 'usa', 'United States', 'united states', 'America', 'america', '🇺🇸')
   OR (
      region = ''
      AND (
        raw_name LIKE '%美国%' OR raw_name LIKE '%美西%' OR raw_name LIKE '%美东%' OR raw_name LIKE '%旧金山%' OR raw_name LIKE '%华盛顿%' OR raw_name LIKE '%维加斯%' OR raw_name LIKE '%🇺🇸%'
        OR display_name LIKE '%美国%' OR display_name LIKE '%美西%' OR display_name LIKE '%美东%' OR display_name LIKE '%旧金山%' OR display_name LIKE '%华盛顿%' OR display_name LIKE '%维加斯%' OR display_name LIKE '%🇺🇸%'
      )
   );

UPDATE upstream_nodes
SET region = '加拿大',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('加拿大', 'Canada', 'canada', '🇨🇦')
   OR (region = '' AND (raw_name LIKE '%加拿大%' OR raw_name LIKE '%🇨🇦%' OR display_name LIKE '%加拿大%' OR display_name LIKE '%🇨🇦%'));

UPDATE upstream_nodes
SET region = '土耳其',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('土耳其', 'Turkey', 'turkey', '🇹🇷')
   OR (region = '' AND (raw_name LIKE '%土耳其%' OR raw_name LIKE '%🇹🇷%' OR display_name LIKE '%土耳其%' OR display_name LIKE '%🇹🇷%'));

UPDATE upstream_nodes
SET region = '俄罗斯',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('俄罗斯', 'Russia', 'russia', '🇷🇺')
   OR (region = '' AND (raw_name LIKE '%俄罗斯%' OR raw_name LIKE '%🇷🇺%' OR display_name LIKE '%俄罗斯%' OR display_name LIKE '%🇷🇺%'));

UPDATE upstream_nodes
SET region = '越南',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('越南', 'Vietnam', 'vietnam', '🇻🇳')
   OR (region = '' AND (raw_name LIKE '%越南%' OR raw_name LIKE '%🇻🇳%' OR display_name LIKE '%越南%' OR display_name LIKE '%🇻🇳%'));

UPDATE upstream_nodes
SET region = '印尼',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('印尼', '印度尼西亚', 'Indonesia', 'indonesia', '🇮🇩')
   OR (region = '' AND (raw_name LIKE '%印尼%' OR raw_name LIKE '%印度尼西亚%' OR raw_name LIKE '%🇮🇩%' OR display_name LIKE '%印尼%' OR display_name LIKE '%印度尼西亚%' OR display_name LIKE '%🇮🇩%'));

UPDATE upstream_nodes
SET region = '日本',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('日本', '东京', '大阪', 'JP', 'jp', 'Japan', 'japan', 'Tokyo', 'tokyo', 'Osaka', 'osaka', '🇯🇵')
   OR (region = '' AND (raw_name LIKE '%日本%' OR raw_name LIKE '%东京%' OR raw_name LIKE '%大阪%' OR raw_name LIKE '%🇯🇵%' OR display_name LIKE '%日本%' OR display_name LIKE '%东京%' OR display_name LIKE '%大阪%' OR display_name LIKE '%🇯🇵%'));

UPDATE upstream_nodes
SET region = '韩国',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('韩国', '韓國', '首尔', '首爾', 'KR', 'kr', 'Korea', 'korea', 'Seoul', 'seoul', '🇰🇷')
   OR (region = '' AND (raw_name LIKE '%韩国%' OR raw_name LIKE '%韓國%' OR raw_name LIKE '%首尔%' OR raw_name LIKE '%首爾%' OR raw_name LIKE '%🇰🇷%' OR display_name LIKE '%韩国%' OR display_name LIKE '%韓國%' OR display_name LIKE '%首尔%' OR display_name LIKE '%首爾%' OR display_name LIKE '%🇰🇷%'));

UPDATE upstream_nodes
SET region = '新加坡',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('新加坡', 'SG', 'sg', 'Singapore', 'singapore', '🇸🇬')
   OR (region = '' AND (raw_name LIKE '%新加坡%' OR raw_name LIKE '%🇸🇬%' OR display_name LIKE '%新加坡%' OR display_name LIKE '%🇸🇬%'));

UPDATE upstream_nodes
SET region = '澳洲',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('澳洲', '澳大利亚', '澳大利亞', 'Australia', 'australia', '🇦🇺')
   OR (region = '' AND (raw_name LIKE '%澳洲%' OR raw_name LIKE '%澳大利亚%' OR raw_name LIKE '%澳大利亞%' OR raw_name LIKE '%🇦🇺%' OR display_name LIKE '%澳洲%' OR display_name LIKE '%澳大利亚%' OR display_name LIKE '%澳大利亞%' OR display_name LIKE '%🇦🇺%'));

UPDATE upstream_nodes
SET region = '阿联酋',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('阿联酋', '阿聯酋', '迪拜', 'Dubai', 'dubai', 'UAE', 'uae', '🇦🇪')
   OR (region = '' AND (raw_name LIKE '%阿联酋%' OR raw_name LIKE '%阿聯酋%' OR raw_name LIKE '%迪拜%' OR raw_name LIKE '%🇦🇪%' OR display_name LIKE '%阿联酋%' OR display_name LIKE '%阿聯酋%' OR display_name LIKE '%迪拜%' OR display_name LIKE '%🇦🇪%'));

UPDATE upstream_nodes
SET region = '印度',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('印度', 'India', 'india', '🇮🇳')
   OR (region = '' AND (raw_name LIKE '%印度%' OR raw_name LIKE '%🇮🇳%' OR display_name LIKE '%印度%' OR display_name LIKE '%🇮🇳%'));

UPDATE upstream_nodes
SET region = '德国',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('德国', '德國', 'Germany', 'germany', '🇩🇪')
   OR (region = '' AND (raw_name LIKE '%德国%' OR raw_name LIKE '%德國%' OR raw_name LIKE '%🇩🇪%' OR display_name LIKE '%德国%' OR display_name LIKE '%德國%' OR display_name LIKE '%🇩🇪%'));

UPDATE upstream_nodes
SET region = '英国',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('英国', '英國', 'UK', 'uk', 'GB', 'gb', 'United Kingdom', 'united kingdom', 'Britain', 'britain', '🇬🇧')
   OR (region = '' AND (raw_name LIKE '%英国%' OR raw_name LIKE '%英國%' OR raw_name LIKE '%🇬🇧%' OR display_name LIKE '%英国%' OR display_name LIKE '%英國%' OR display_name LIKE '%🇬🇧%'));

UPDATE upstream_nodes
SET region = '巴西',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('巴西', 'Brazil', 'brazil', '🇧🇷')
   OR (region = '' AND (raw_name LIKE '%巴西%' OR raw_name LIKE '%🇧🇷%' OR display_name LIKE '%巴西%' OR display_name LIKE '%🇧🇷%'));

UPDATE upstream_nodes
SET region = '智利',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('智利', 'Chile', 'chile', '🇨🇱')
   OR (region = '' AND (raw_name LIKE '%智利%' OR raw_name LIKE '%🇨🇱%' OR display_name LIKE '%智利%' OR display_name LIKE '%🇨🇱%'));

UPDATE upstream_nodes
SET region = '法国',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('法国', '法國', 'France', 'france', '🇫🇷')
   OR (region = '' AND (raw_name LIKE '%法国%' OR raw_name LIKE '%法國%' OR raw_name LIKE '%🇫🇷%' OR display_name LIKE '%法国%' OR display_name LIKE '%法國%' OR display_name LIKE '%🇫🇷%'));

UPDATE upstream_nodes
SET region = '墨西哥',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('墨西哥', 'Mexico', 'mexico', '🇲🇽')
   OR (region = '' AND (raw_name LIKE '%墨西哥%' OR raw_name LIKE '%🇲🇽%' OR display_name LIKE '%墨西哥%' OR display_name LIKE '%🇲🇽%'));

UPDATE upstream_nodes
SET region = '荷兰',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('荷兰', '荷蘭', 'Netherlands', 'netherlands', 'Holland', 'holland', '🇳🇱')
   OR (region = '' AND (raw_name LIKE '%荷兰%' OR raw_name LIKE '%荷蘭%' OR raw_name LIKE '%🇳🇱%' OR display_name LIKE '%荷兰%' OR display_name LIKE '%荷蘭%' OR display_name LIKE '%🇳🇱%'));

UPDATE upstream_nodes
SET region = '以色列',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('以色列', 'Israel', 'israel', '🇮🇱')
   OR (region = '' AND (raw_name LIKE '%以色列%' OR raw_name LIKE '%🇮🇱%' OR display_name LIKE '%以色列%' OR display_name LIKE '%🇮🇱%'));

UPDATE upstream_nodes
SET region = '西班牙',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('西班牙', 'Spain', 'spain', '🇪🇸')
   OR (region = '' AND (raw_name LIKE '%西班牙%' OR raw_name LIKE '%🇪🇸%' OR display_name LIKE '%西班牙%' OR display_name LIKE '%🇪🇸%'));

UPDATE upstream_nodes
SET region = '马来西亚',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('马来西亚', '馬來西亞', 'Malaysia', 'malaysia', '🇲🇾')
   OR (region = '' AND (raw_name LIKE '%马来西亚%' OR raw_name LIKE '%馬來西亞%' OR raw_name LIKE '%🇲🇾%' OR display_name LIKE '%马来西亚%' OR display_name LIKE '%馬來西亞%' OR display_name LIKE '%🇲🇾%'));

UPDATE upstream_nodes
SET region = '菲律宾',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('菲律宾', '菲律賓', 'Philippines', 'philippines', '🇵🇭')
   OR (region = '' AND (raw_name LIKE '%菲律宾%' OR raw_name LIKE '%菲律賓%' OR raw_name LIKE '%🇵🇭%' OR display_name LIKE '%菲律宾%' OR display_name LIKE '%菲律賓%' OR display_name LIKE '%🇵🇭%'));

UPDATE upstream_nodes
SET region = '泰国',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('泰国', '泰國', 'Thailand', 'thailand', '🇹🇭')
   OR (region = '' AND (raw_name LIKE '%泰国%' OR raw_name LIKE '%泰國%' OR raw_name LIKE '%🇹🇭%' OR display_name LIKE '%泰国%' OR display_name LIKE '%泰國%' OR display_name LIKE '%🇹🇭%'));

UPDATE upstream_nodes
SET region = '阿根廷',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('阿根廷', 'Argentina', 'argentina', '🇦🇷')
   OR (region = '' AND (raw_name LIKE '%阿根廷%' OR raw_name LIKE '%🇦🇷%' OR display_name LIKE '%阿根廷%' OR display_name LIKE '%🇦🇷%'));

UPDATE upstream_nodes
SET region = '乌克兰',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('乌克兰', '烏克蘭', 'Ukraine', 'ukraine', '🇺🇦')
   OR (region = '' AND (raw_name LIKE '%乌克兰%' OR raw_name LIKE '%烏克蘭%' OR raw_name LIKE '%🇺🇦%' OR display_name LIKE '%乌克兰%' OR display_name LIKE '%烏克蘭%' OR display_name LIKE '%🇺🇦%'));

UPDATE upstream_nodes
SET region = '瑞士',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('瑞士', 'Switzerland', 'switzerland', '🇨🇭')
   OR (region = '' AND (raw_name LIKE '%瑞士%' OR raw_name LIKE '%🇨🇭%' OR display_name LIKE '%瑞士%' OR display_name LIKE '%🇨🇭%'));

UPDATE upstream_nodes
SET region = '南非',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('南非', 'South Africa', 'south africa', 'Johannesburg', 'johannesburg', '约翰内斯堡', '約翰內斯堡', '🇿🇦')
   OR (region = '' AND (raw_name LIKE '%南非%' OR raw_name LIKE '%约翰内斯堡%' OR raw_name LIKE '%約翰內斯堡%' OR raw_name LIKE '%🇿🇦%' OR display_name LIKE '%南非%' OR display_name LIKE '%约翰内斯堡%' OR display_name LIKE '%約翰內斯堡%' OR display_name LIKE '%🇿🇦%'));

UPDATE upstream_nodes
SET region = '🇨🇳中国|香港',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('香港', 'HK', 'hk', 'HKG', 'hkg', 'Hong Kong', 'hong kong', 'HongKong', 'hongkong', '🇭🇰')
   OR (
      region = ''
      AND (
        raw_name LIKE '%香港%' OR raw_name LIKE '%🇭🇰%' OR raw_name LIKE '%Hong Kong%' OR raw_name LIKE '%hong kong%' OR raw_name LIKE '%HongKong%' OR raw_name LIKE '%hongkong%'
        OR display_name LIKE '%香港%' OR display_name LIKE '%🇭🇰%' OR display_name LIKE '%Hong Kong%' OR display_name LIKE '%hong kong%' OR display_name LIKE '%HongKong%' OR display_name LIKE '%hongkong%'
      )
   );

UPDATE upstream_nodes
SET region = '🇨🇳中国|台湾',
    updated_at = CURRENT_TIMESTAMP
WHERE region IN ('台湾', '台灣', '台北', 'TW', 'tw', 'TPE', 'tpe', 'Taiwan', 'taiwan', 'Taipei', 'taipei', '🇹🇼')
   OR (
      region = ''
      AND (
        raw_name LIKE '%台湾%' OR raw_name LIKE '%台灣%' OR raw_name LIKE '%台北%' OR raw_name LIKE '%🇹🇼%' OR raw_name LIKE '%Taiwan%' OR raw_name LIKE '%taiwan%' OR raw_name LIKE '%Taipei%' OR raw_name LIKE '%taipei%'
        OR display_name LIKE '%台湾%' OR display_name LIKE '%台灣%' OR display_name LIKE '%台北%' OR display_name LIKE '%🇹🇼%' OR display_name LIKE '%Taiwan%' OR display_name LIKE '%taiwan%' OR display_name LIKE '%Taipei%' OR display_name LIKE '%taipei%'
      )
   );
