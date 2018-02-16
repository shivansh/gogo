	.data
nStr:	.asciiz "Enter n: "
n:	.word	0
count:	.word	0
str:	.asciiz "Number of digits in n: "

	.text


	.globl main
	.ent main
main:
	li $v0, 4
	la $a0, nStr
	syscall
	li $v0, 5
	syscall
	move $t1, $v0
	li $t4, 0		# count -> $t4
	# Store variables back into memory
	sw $t1, n
	sw $t4, count

while:

	lw $t1, n
	ble $t1, 0, exit		# exit -> $t0
	div $t1, $t1, 10		# n -> $t1
	lw $t4, count
	addi $t4, $t4, 1		# count -> $t4
	# Store variables back into memory
	sw $t1, n
	sw $t4, count

	j while

exit:
	li $v0, 4
	la $a0, str
	syscall
	li $v0, 1
	lw $t1, count
	move $a0, $t1
	syscall
	# Store variables back into memory
	sw $t1, count
	li $v0, 10
	syscall
	.end main
