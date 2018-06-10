# Test to declare and print a string

	.data
str:		.asciiz "Hello World!"

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime


	.globl main
	.ent main
main:
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 10
	syscall
	.end main
