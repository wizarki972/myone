package common

const CONFIG_DIR = ".config/myone"
const CACHE_DIR = ".cache/myone"
const BASE_DIR = ".local/share/myone"

const SCRIPTS_DIR = BASE_DIR + "/scripts" // do I really needs this const?! && do I need to move the scripts folder

// WALLS
const ALL_WALLS_DIR = BASE_DIR + "/walls"

const CURRENT_WALLPAPER_ENTRY = ".current-wallpaper"
const CURRENT_WALLPAPER_ENTRY_PATH = BASE_DIR + "/" + CURRENT_WALLPAPER_ENTRY

// THEMES
const THEMES_DIR = BASE_DIR + "/themes"

const CURRENT_THEME_NAME_ENTRY = ".current-theme"
const CURRENT_THEME_NAME_ENTRY_PATH = BASE_DIR + "/" + CURRENT_THEME_NAME_ENTRY

const THEMES_STATE_DIR = ".local/state/myone"
const COMMON_PLACED_STATE_PATH = THEMES_STATE_DIR + "/is_common_placed"
