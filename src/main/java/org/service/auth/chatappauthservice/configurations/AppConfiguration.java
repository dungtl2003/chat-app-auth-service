package org.service.auth.chatappauthservice.configurations;

import org.service.auth.chatappauthservice.utils.UserDTOMapper;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;

@Configuration
public class AppConfiguration {

	@Bean
	public UserDTOMapper userDTOMapper() {
		return new UserDTOMapper();
	}

	@Bean
	public PasswordEncoder passwordEncoder() {
		return new BCryptPasswordEncoder(Integer.parseInt(System.getenv("BCRYPT_STRENGTH")));
	}

}
