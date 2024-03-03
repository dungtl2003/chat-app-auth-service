package org.service.auth.chatappauthservice.entity.enums;

public enum Gender {

	MALE("male"), FEMALE("female"), OTHER("other"),;

	private final String name;

	Gender(String name) {
		this.name = name;
	}

	public String getName() {
		return this.name;
	}

}
