import { useDispatch, useSelector } from 'react-redux';
import { SET_THEME } from 'store/actions';
import { useTheme } from '@mui/material/styles';
import { Avatar, Box, ButtonBase, Tooltip } from '@mui/material';
import { Icon } from '@iconify/react';
import { useTranslation } from 'react-i18next';

export default function ThemeButton() {
  const dispatch = useDispatch();
  const { t } = useTranslation();

  const defaultTheme = useSelector((state) => state.customization.theme);

  const theme = useTheme();

  // 获取当前主题模式：auto（跟随系统）、light、dark
  const getThemeMode = () => {
    const storedTheme = localStorage.getItem('theme');
    if (!storedTheme) return 'auto';
    return storedTheme;
  };

  // 获取显示的图标和提示文字
  const getThemeDisplay = () => {
    const mode = getThemeMode();
    switch (mode) {
      case 'auto':
        return {
          icon: 'solar:monitor-bold-duotone',
          tooltip: t('theme.auto')
        };
      case 'light':
        return {
          icon: 'solar:sun-2-bold-duotone',
          tooltip: t('theme.light')
        };
      case 'dark':
        return {
          icon: 'solar:moon-bold-duotone',
          tooltip: t('theme.dark')
        };
      default:
        return {
          icon: 'solar:monitor-bold-duotone',
          tooltip: t('theme.auto')
        };
    }
  };

  const handleThemeChange = () => {
    const currentMode = getThemeMode();
    let nextMode;
    let nextTheme;

    // 循环切换：auto → light → dark → auto
    switch (currentMode) {
      case 'auto':
        nextMode = 'light';
        nextTheme = 'light';
        localStorage.setItem('theme', 'light');
        break;
      case 'light':
        nextMode = 'dark';
        nextTheme = 'dark';
        localStorage.setItem('theme', 'dark');
        break;
      case 'dark':
        nextMode = 'auto';
        // 跟随系统时，移除localStorage中的设置
        localStorage.removeItem('theme');
        // 检测当前系统主题
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        nextTheme = prefersDark ? 'dark' : 'light';
        break;
      default:
        nextMode = 'light';
        nextTheme = 'light';
        localStorage.setItem('theme', 'light');
    }

    dispatch({ type: SET_THEME, theme: nextTheme });
  };

  const { icon, tooltip } = getThemeDisplay();

  return (
    <Box
      sx={{
        ml: 2,
        mr: 3,
        [theme.breakpoints.down('md')]: {
          mr: 2
        }
      }}
    >
      <Tooltip title={tooltip} placement="bottom">
        <ButtonBase sx={{ borderRadius: '12px' }}>
          <Avatar
            variant="rounded"
            sx={{
              ...theme.typography.commonAvatar,
              ...theme.typography.mediumAvatar,
              ...theme.typography.menuButton,
              transition: 'all .2s ease-in-out',
              borderColor: 'transparent',
              backgroundColor: 'transparent',
              boxShadow: 'none',
              borderRadius: '50%',
              '&[aria-controls="menu-list-grow"],&:hover': {
                boxShadow: '0 0 10px rgba(0,0,0,0.2)',
                backgroundColor: 'transparent',
                borderRadius: '50%'
              }
            }}
            onClick={handleThemeChange}
            color="inherit"
          >
            <Icon icon={icon} width="1.5rem" />
          </Avatar>
        </ButtonBase>
      </Tooltip>
    </Box>
  );
}
