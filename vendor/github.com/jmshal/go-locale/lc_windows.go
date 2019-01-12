package go_locale

import "errors"

type LC struct {
    index map[string]int
}

var UNKNOWN_LOCALE error = errors.New("unknown locale")
var UNKNOWN_ID error = errors.New("unknown id")

var _lc LC

func LCID() LC {
    return _lc
}

func (lc *LC) ByLocaleString(localeString string) (int, error) {
    for lcidLocaleString, lcidId := range lc.index {
        if localeString == lcidLocaleString {
            return lcidId, nil
        }
    }
    return 0, UNKNOWN_LOCALE
}

func (lc *LC) ById(id int) (string, error) {
    for lcidLocaleString, lcidId := range lc.index {
        if id == lcidId {
            return lcidLocaleString, nil
        }
    }
    return "", UNKNOWN_ID
}

func init() {
    // Source: https://raw.githubusercontent.com/sindresorhus/lcid/master/lcid.json
    index := make(map[string]int, 201)
    index["af_ZA"] = 1078
    index["am_ET"] = 1118
    index["ar_AE"] = 14337
    index["ar_BH"] = 15361
    index["ar_DZ"] = 5121
    index["ar_EG"] = 3073
    index["ar_IQ"] = 2049
    index["ar_JO"] = 11265
    index["ar_KW"] = 13313
    index["ar_LB"] = 12289
    index["ar_LY"] = 4097
    index["ar_MA"] = 6145
    index["ar_OM"] = 8193
    index["ar_QA"] = 16385
    index["ar_SA"] = 1025
    index["ar_SY"] = 10241
    index["ar_TN"] = 7169
    index["ar_YE"] = 9217
    index["arn_CL"] = 1146
    index["as_IN"] = 1101
    index["az_AZ"] = 2092
    index["ba_RU"] = 1133
    index["be_BY"] = 1059
    index["bg_BG"] = 1026
    index["bn_IN"] = 1093
    index["bo_BT"] = 2129
    index["bo_CN"] = 1105
    index["br_FR"] = 1150
    index["bs_BA"] = 8218
    index["ca_ES"] = 1027
    index["co_FR"] = 1155
    index["cs_CZ"] = 1029
    index["cy_GB"] = 1106
    index["da_DK"] = 1030
    index["de_AT"] = 3079
    index["de_CH"] = 2055
    index["de_DE"] = 1031
    index["de_LI"] = 5127
    index["de_LU"] = 4103
    index["div_MV"] = 1125
    index["dsb_DE"] = 2094
    index["el_GR"] = 1032
    index["en_AU"] = 3081
    index["en_BZ"] = 10249
    index["en_CA"] = 4105
    index["en_CB"] = 9225
    index["en_GB"] = 2057
    index["en_IE"] = 6153
    index["en_IN"] = 18441
    index["en_JA"] = 8201
    index["en_MY"] = 17417
    index["en_NZ"] = 5129
    index["en_PH"] = 13321
    index["en_TT"] = 11273
    index["en_US"] = 1033
    index["en_ZA"] = 7177
    index["en_ZW"] = 12297
    index["es_AR"] = 11274
    index["es_BO"] = 16394
    index["es_CL"] = 13322
    index["es_CO"] = 9226
    index["es_CR"] = 5130
    index["es_DO"] = 7178
    index["es_EC"] = 12298
    index["es_ES"] = 3082
    index["es_GT"] = 4106
    index["es_HN"] = 18442
    index["es_MX"] = 2058
    index["es_NI"] = 19466
    index["es_PA"] = 6154
    index["es_PE"] = 10250
    index["es_PR"] = 20490
    index["es_PY"] = 15370
    index["es_SV"] = 17418
    index["es_UR"] = 14346
    index["es_US"] = 21514
    index["es_VE"] = 8202
    index["et_EE"] = 1061
    index["eu_ES"] = 1069
    index["fa_IR"] = 1065
    index["fi_FI"] = 1035
    index["fil_PH"] = 1124
    index["fo_FO"] = 1080
    index["fr_BE"] = 2060
    index["fr_CA"] = 3084
    index["fr_CH"] = 4108
    index["fr_FR"] = 1036
    index["fr_LU"] = 5132
    index["fr_MC"] = 6156
    index["fy_NL"] = 1122
    index["ga_IE"] = 2108
    index["gbz_AF"] = 1164
    index["gl_ES"] = 1110
    index["gsw_FR"] = 1156
    index["gu_IN"] = 1095
    index["ha_NG"] = 1128
    index["he_IL"] = 1037
    index["hi_IN"] = 1081
    index["hr_BA"] = 4122
    index["hr_HR"] = 1050
    index["hu_HU"] = 1038
    index["hy_AM"] = 1067
    index["id_ID"] = 1057
    index["ii_CN"] = 1144
    index["is_IS"] = 1039
    index["it_CH"] = 2064
    index["it_IT"] = 1040
    index["iu_CA"] = 2141
    index["ja_JP"] = 1041
    index["ka_GE"] = 1079
    index["kh_KH"] = 1107
    index["kk_KZ"] = 1087
    index["kl_GL"] = 1135
    index["kn_IN"] = 1099
    index["ko_KR"] = 1042
    index["kok_IN"] = 1111
    index["ky_KG"] = 1088
    index["lb_LU"] = 1134
    index["lo_LA"] = 1108
    index["lt_LT"] = 1063
    index["lv_LV"] = 1062
    index["mi_NZ"] = 1153
    index["mk_MK"] = 1071
    index["ml_IN"] = 1100
    index["mn_CN"] = 2128
    index["mn_MN"] = 1104
    index["moh_CA"] = 1148
    index["mr_IN"] = 1102
    index["ms_BN"] = 2110
    index["ms_MY"] = 1086
    index["mt_MT"] = 1082
    index["my_MM"] = 1109
    index["nb_NO"] = 1044
    index["ne_NP"] = 1121
    index["nl_BE"] = 2067
    index["nl_NL"] = 1043
    index["nn_NO"] = 2068
    index["ns_ZA"] = 1132
    index["oc_FR"] = 1154
    index["or_IN"] = 1096
    index["pa_IN"] = 1094
    index["pl_PL"] = 1045
    index["ps_AF"] = 1123
    index["pt_BR"] = 1046
    index["pt_PT"] = 2070
    index["qut_GT"] = 1158
    index["quz_BO"] = 1131
    index["quz_EC"] = 2155
    index["quz_PE"] = 3179
    index["rm_CH"] = 1047
    index["ro_RO"] = 1048
    index["ru_RU"] = 1049
    index["rw_RW"] = 1159
    index["sa_IN"] = 1103
    index["sah_RU"] = 1157
    index["se_FI"] = 3131
    index["se_NO"] = 1083
    index["se_SE"] = 2107
    index["si_LK"] = 1115
    index["sk_SK"] = 1051
    index["sl_SI"] = 1060
    index["sma_NO"] = 6203
    index["sma_SE"] = 7227
    index["smj_NO"] = 4155
    index["smj_SE"] = 5179
    index["smn_FI"] = 9275
    index["sms_FI"] = 8251
    index["sq_AL"] = 1052
    index["sr_BA"] = 7194
    index["sr_SP"] = 3098
    index["sv_FI"] = 2077
    index["sv_SE"] = 1053
    index["sw_KE"] = 1089
    index["syr_SY"] = 1114
    index["ta_IN"] = 1097
    index["te_IN"] = 1098
    index["tg_TJ"] = 1064
    index["th_TH"] = 1054
    index["tk_TM"] = 1090
    index["tmz_DZ"] = 2143
    index["tn_ZA"] = 1074
    index["tr_TR"] = 1055
    index["tt_RU"] = 1092
    index["ug_CN"] = 1152
    index["uk_UA"] = 1058
    index["ur_IN"] = 2080
    index["ur_PK"] = 1056
    index["uz_UZ"] = 2115
    index["vi_VN"] = 1066
    index["wen_DE"] = 1070
    index["wo_SN"] = 1160
    index["xh_ZA"] = 1076
    index["yo_NG"] = 1130
    index["zh_CHS"] = 4
    index["zh_CHT"] = 31748
    index["zh_CN"] = 2052
    index["zh_HK"] = 3076
    index["zh_MO"] = 5124
    index["zh_SG"] = 4100
    index["zh_TW"] = 1028
    index["zu_ZA"] = 1077
    _lc = LC{index}
}
