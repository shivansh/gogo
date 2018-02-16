# Test to declare and print a string

	.data
str:	.asciiz "Hello World!"

	.text


	.globl main
	.ent main
main:
	li $v0, 4
	la $a0, str
	syscall
	li $v0, 10
	syscall
	.end main
