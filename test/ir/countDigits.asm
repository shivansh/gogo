	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
count:		.word	0
str:		.asciiz "Number of digits in n: "

	.text
	.globl main
	.ent main
main:
	li	$v0, 4
	la	$a0, nStr
	syscall
	li	$v0, 5
	syscall
	move	$3, $v0
	li	$7, 0		# count -> $7
	# Store variables back into memory
	sw	$3, n
	sw	$7, count

while:
	lw	$3, n
	ble	$3, 0, exit	# exit -> $0
	div	$3, $3, 10	# n -> $3
	lw	$7, count
	addi	$7, $7, 1	# count -> $7
	# Store variables back into memory
	sw	$3, n
	sw	$7, count
	j	while

exit:
	li	$v0, 4
	la	$a0, str
	syscall
	li	$v0, 1
	lw	$3, count
	move	$a0, $3
	syscall
	# Store variables back into memory
	sw	$3, count
	li	$v0, 10
	syscall
	.end main
