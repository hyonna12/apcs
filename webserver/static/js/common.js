// Ajax
function jcall(method, url, params, callback) {
    $.ajax({
        cache: false,
        url: url,
        dataType: 'json',
        type: method,
        data: params,
        success: function success(data) {
            callback(null, data);
        },
        error: function error(xhr, status, _error2) {
            callback(_error2);
        }
    });
}

// localStorage 아이템 추가
// 모든 key와 value는 항상 string으로 저장됩니다. (주의하세요, object와 integer 들은 string으로 자동 변환됩니다.)
function setLocalStorage(key, value) {
	localStorage.setItem(key, value);
}

// localStorage 아이템 읽기
// return값은 항상 string입니다.
function getLocalStorage(key) {
	return localStorage.getItem(key);
}

// localStorage 아이템 삭제
function removeLocalStorage(key) {
	localStorage.removeItem(key);
}

// localStorage 아이템 전체 삭제
function clearLocalStorage() {
	localStorage.clear();
}

// URL Parameter 만들기 
async function makeParamsUrl(data){
    let url_params = '';
    await Object.keys(data).reduce(async (pre, key, index) => {
        return pre.then(async ()=> new Promise(async (resolve) => {
            url_params += (index === 0 ? "?" : "&") + key + "=" + String(data[key]);
            resolve();
        }));
    }, Promise.resolve());
    return url_params;
}

// 모바일 체크 여부
function isMobile() { 
    var user = navigator.userAgent; 
    var is_mobile = false; 
    if( user.indexOf("iPhone") > -1 || user.indexOf("Android") > -1 ) { 
        is_mobile = true; 
    } 
    return is_mobile; 
}

// 쿠키 관련
/* 
    name : 쿠키 키 값
    value : 값
    exp : 시간
*/
function setCookie(name, value, exp) {
    var date = new Date();
    date.setTime(date.getTime() + exp * 24 * 60 * 60 * 1000);
    document.cookie = name + '=' + value + ';expires=' + date.toUTCString() + ';path=/';
};
function getCookie(name) {
    var value = document.cookie.match('(^|;) ?' + name + '=([^;]*)(;|$)');
    return value? value[2] : null;
};
function deleteCookie(name) {    
    document.cookie = name + '=; expires=Thu, 01 Jan 1999 00:00:10 GMT;';
};
// 쿠키 관련

// 콤마
function setComma(n){
    var reg = /(^[+-]?\d+)(\d{3})/;   // 정규식
    n += '';                          // 숫자를 문자열로 변환         
    while (reg.test(n)) {
        n = n.replace(reg, '$1' + ',' + '$2');
    }         
    return n;
};

// 사업자 번호 포맷
function bizNoFormatter(num, type) {

    var formatNum = '';

    try{
        if (num.length == 10) {
            if (type == 0) {
                formatNum = num.replace(/(\d{3})(\d{2})(\d{5})/, '$1-$2-*****');
            } else {
                formatNum = num.replace(/(\d{3})(\d{2})(\d{5})/, '$1-$2-$3');
            }
        }
    } catch(e) {
        formatNum = num;
        console.log(e);
    }
    return formatNum;
}

// 한글로 금액 구하기
/* 
    obj => 금액
*/
function industryCalcul(obj){
    const formatter = Intl.NumberFormat();
    
    if(parseInt(obj) > 9999999999999999 || parseInt(obj) < -9999999999999999) {
        var gung = String(obj).slice(0,-16);
        obj = (obj % 10000000000000000);
        var jo = (obj / 1000000000000).toFixed(0);

        if(gung == 0){
            return formatter.format(jo) + '조';
        }else{
            if(parseInt(jo) < 0){
                jo = jo * -1;
            }
            return formatter.format(gung) + '경 ' + formatter.format(jo) + '조 ';
        }
    }
    else if(parseInt(obj) > 999999999999 || parseInt(obj) < -999999999999) {
        var jo = String(obj).slice(0,-12);
        obj = (obj % 1000000000000);
        var eok = (obj / 100000000).toFixed(0);

        if(jo == 0){
            return formatter.format(eok) + '억';
        }else{
            if(parseInt(eok) < 0){
                eok = eok * -1;
            }
            return formatter.format(jo) + '조 ' + formatter.format(eok) + '억';
        }
    } else if (parseInt(obj) > 99999999 || parseInt(obj) < -99999999) {
        obj = (obj / 100000000).toFixed(0);
        return formatter.format(obj) + '억';
    } else {
        return formatter.format(obj);
    }
}