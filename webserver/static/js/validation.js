$(function(){
    // 전화번호 keyup
    $(document).on("keyup", "#tel", function() { 
        $(this).val( $(this).val().replace(/[^0-9]/g, "").replace(/(^02|^0505|^1[0-9]{3}|^0[0-9]{2})([0-9]+)?([0-9]{4})$/,"$1-$2-$3").replace("--", "-") ); 
    });
})

// 이메일 형식 체크
function doCheckEmail(email){
    let emailRegExp = /^[A-Za-z0-9_]+[A-Za-z0-9]*[@]{1}[A-Za-z0-9_-]+[A-Za-z0-9_-]*[.]{1}[A-Za-z]{1,3}$/;
    if (!emailRegExp.test(email)) {
        alert('이메일 형식이 올바르지 않습니다.');
        return false;
    }else{
        return true;
    }
}

// 비밀번호 체크
function doCheckPassword(){
    if($('#cur_password').val() == ''){
        alert('현재 비밀번호를 입력하세요.');
        $('#cur_password').focus();
        return false;
    }
    if($('#new_password').val() == ''){
        alert('새 비밀번호를 입력하세요.');
        $('#new_password').focus();
        return false;
    }
    if($('#new_password_confirm').val() == ''){
        alert('비밀번호 확인을 입력하세요.');
        $('#new_password_confirm').focus();
        return false;
    }
    if($('#cur_password').val() == $('#new_password').val()){
        alert('현재 비밀번호와 새 비밀번호가 같습니다. \n확인해주세요.');
        return false;
    }
    if($('#new_password').val() != $('#new_password_confirm').val()){
        alert('비밀번호가 서로 일치하지 않습니다. \n확인해주세요.');
        return false;
    }
    let pass =  $('#new_password').val();
    if(pass.length < 4 || pass.length > 26){
        alert('비밀번호를 다시 확인해주세요. \n(4~26/영문/숫자 가능)');
        $('#new_password').val('');
        $('#new_password_confirm').val('');
        return false;
    }

    return true;
}